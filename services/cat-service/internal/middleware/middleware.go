package middleware

import (
	"errors"
	"github.com/chocological13/kittykeeper/cat-service/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type CatOwnershipChecker interface {
	VerifyCatOwnership(ctx *gin.Context, userID, catID uuid.UUID) (bool, error)
}

type AuthMiddleware struct {
	jwtService *auth.JWTService
	tokenStore *auth.TokenStore
	log        *log.Entry
}

func NewAuthMiddleware(jwtService *auth.JWTService, tokenStore *auth.TokenStore, log *log.Entry) *AuthMiddleware {
	log = log.WithField("middleware", "auth")
	return &AuthMiddleware{
		jwtService: jwtService,
		tokenStore: tokenStore,
		log:        log,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Error("authorization header is missing")
			c.AbortWithStatusJSON(401, gin.H{"error": "authorization header is missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Error("invalid authorization header format")
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtService.VerifyToken(tokenString)
		if err != nil {
			var statusCode int
			if errors.Is(err, auth.ErrExpiredToken) {
				log.Error("expired token")
				statusCode = http.StatusUnauthorized
			} else if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrInvalidTokenClaims) {
				log.Error("invalid token")
				statusCode = http.StatusBadRequest
			} else {
				log.Error("unknown error")
				statusCode = http.StatusUnauthorized
			}
			c.AbortWithStatusJSON(statusCode, gin.H{"error": err.Error()})
			return
		}

		userID := claims.UserID

		storedToken, err := m.tokenStore.GetAccessToken(c.Request.Context(), userID)
		if err != nil {
			log.Error("failed to get access token from Redis")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if storedToken != tokenString {
			log.Error("invalid token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func (m *AuthMiddleware) OwnershipCheck(catService CatOwnershipChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// * Get both cat id and user id
		catID, err := uuid.Parse(c.Param("catID"))
		if err != nil {
			log.Error("invalid cat id")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid cat id"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			log.Error("user not authenticated")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// * Check ownership
		isOwner, err := catService.VerifyCatOwnership(c, userID.(uuid.UUID), catID)
		if err != nil {
			log.Error("failed to verify cat ownership")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !isOwner {
			log.Warnf("user is not owner of cat with id: %s", catID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user is not owner of cat"})
			return
		}

		c.Next()
	}
}
