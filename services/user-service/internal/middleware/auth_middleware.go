package middleware

import (
	"errors"
	"github.com/chocological13/kittykeeper/services/user-service/internal/auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	authService *auth.AuthService
	tokenStore  *auth.TokenStore
}

func NewAuthMiddleware(authService *auth.AuthService, tokenStore *auth.TokenStore) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		tokenStore:  tokenStore,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// * Get the authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is missing"})
			return
		}

		// * Check the authorization header format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		// * Extract bearer token
		tokenString := parts[1]

		// * Verify the access token
		claims, err := m.authService.VerifyToken(tokenString, auth.AccessToken)
		if err != nil {
			var statusCode int
			if errors.Is(err, auth.ErrExpiredToken) {
				statusCode = http.StatusUnauthorized
			} else {
				statusCode = http.StatusForbidden
			}
			c.AbortWithStatusJSON(statusCode, gin.H{"error": err.Error()})
			return
		}

		// * Verify the token exists in Redis
		storedToken, err := m.tokenStore.GetToken(c.Request.Context(), claims.UserID, auth.AccessToken)
		if err != nil {
			if errors.Is(err, auth.ErrTokenNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if storedToken != tokenString {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// * Set the user ID in the context
		c.Set("userID", claims.UserID)

		// Call the next handler
		c.Next()
	}
}
