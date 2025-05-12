package handlers

import (
	"errors"
	"github.com/chocological13/kittykeeper/services/user-service/internal/models"
	"github.com/chocological13/kittykeeper/services/user-service/internal/service"
	"github.com/chocological13/kittykeeper/services/user-service/internal/utils"
	"github.com/gin-gonic/gin"
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

func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs := utils.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errs})
		return
	}

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
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

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
		errs := utils.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errs})
		return
	}

	user, tokenPair, err := h.userService.Login(c.Request.Context(), models.LoginParams{
		Credential: req.Credential,
		Password:   req.Password,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidCredentials) {
			statusCode = http.StatusUnauthorized
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

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
