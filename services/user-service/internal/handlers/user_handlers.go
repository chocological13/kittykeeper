package handlers

import (
	"errors"
	"github.com/chocological13/kittykeeper/services/user-service/internal/auth"
	"github.com/chocological13/kittykeeper/services/user-service/internal/logger"
	"github.com/chocological13/kittykeeper/services/user-service/internal/models"
	"github.com/chocological13/kittykeeper/services/user-service/internal/service"
	"github.com/chocological13/kittykeeper/services/user-service/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

var log = logger.NewLogger("user-service")

func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("failed to bind request to register user")
		errs := utils.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errs})
		return
	}

	log.Info("Registering new user")
	user, err := h.userService.RegisterUser(c.Request.Context(), models.CreateUserParams{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrEmailAlreadyExists) || errors.Is(err, service.ErrUsernameAlreadyExists) {
			log.WithError(err).Warnf("failed to register user with username: %s, email: %s", req.Username, req.Email)
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	log.Info("Successfully registered new user!")
	c.JSON(http.StatusCreated, models.RegisterResponse{
		User: models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("failed to bind request to login user")
		errs := utils.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errs})
		return
	}

	log.Infof("Logging in user")
	user, tokenPair, err := h.userService.Login(c.Request.Context(), models.LoginParams{
		Credential: req.Credential,
		Password:   req.Password,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidCredentials) {
			log.WithError(err).Error("failed to login user")
			statusCode = http.StatusUnauthorized
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	log.Infof("Successfully logged in user with id: %s", user.ID)
	c.JSON(http.StatusOK, models.LoginResponse{
		User: models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	// * Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		log.Warnf("user not authenticated")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		log.Error("invalid user id")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	log.Infof("Getting user profile for user with id: %s", id)
	user, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrUserNotFound) {
			log.WithError(err).Warnf("user not found with id: %s", id)
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	log.Infof("Successfully got user profile for user with id: %s", id)
	c.JSON(http.StatusOK, gin.H{
		"user": models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("failed to bind request")
		errs := utils.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errs})
		return
	}

	log.Debug("Refreshing tokens for user")
	newTokens, err := h.userService.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrExpiredToken) {
			log.WithError(err).Warn("failed to refresh tokens")
			statusCode = http.StatusUnauthorized
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	log.Debug("Successfully refreshed tokens for user")
	c.JSON(http.StatusOK, models.RefreshTokenResponse{
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	// * Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		log.Warn("user not authenticated")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		log.Error("invalid user id")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.userService.Logout(c.Request.Context(), id); err != nil {
		log.WithError(err).Errorf("failed to logout user with id: %s", id)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Infof("Successfully logged out user with id: %s", id)
	c.Status(http.StatusNoContent)
}
