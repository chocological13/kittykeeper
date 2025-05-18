package middleware

import (
	"context"
	"errors"
	"github.com/chocological13/kittykeeper/cat-service/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type CatOwnershipChecker interface {
	VerifyCatOwnership(ctx context.Context, catID, userID uuid.UUID) (bool, error)
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
			m.log.Error("authorization header is missing")
			c.AbortWithStatusJSON(401, gin.H{"error": "authorization header is missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.log.Error("invalid authorization header format")
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtService.VerifyToken(tokenString)
		if err != nil {
			var statusCode int
			if errors.Is(err, auth.ErrExpiredToken) {
				m.log.WithError(err).Error("token verification failed: expired token")
				statusCode = http.StatusUnauthorized
			} else if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrInvalidTokenClaims) {
				m.log.WithError(err).Error("token verification failed: invalid token")
				statusCode = http.StatusBadRequest
			} else {
				m.log.WithError(err).Error("token verification failed: unknown error")
				statusCode = http.StatusUnauthorized
			}
			c.AbortWithStatusJSON(statusCode, gin.H{"error": err.Error()})
			return
		}

		userID := claims.UserID

		storedToken, err := m.tokenStore.GetAccessToken(c.Request.Context(), userID)
		if err != nil {
			m.log.WithError(err).Error("failed to get access token from Redis")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if storedToken != tokenString {
			m.log.WithFields(log.Fields{
				"user_id": userID,
			}).Error("stored token doesn't match provided token")
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
		rawID := c.Param("id")
		catID, err := uuid.Parse(rawID)
		if err != nil {
			m.log.WithError(err).Error("invalid cat id")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid cat id"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			m.log.WithError(err).Error("user not authenticated - userID not found in context")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// * Check ownership
		isOwner, err := catService.VerifyCatOwnership(c.Request.Context(), catID, userID.(uuid.UUID))
		if err != nil {
			m.log.WithFields(log.Fields{
				"cat_id":  catID,
				"user_id": userID,
				"error":   err.Error(),
			}).Error("failed to verify cat ownership")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !isOwner {
			m.log.WithFields(log.Fields{
				"cat_id":  catID,
				"user_id": userID,
			}).Warn("user attempted to access cat they don't own")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user is not owner of cat"})
			return
		}

		c.Next()
	}
}
