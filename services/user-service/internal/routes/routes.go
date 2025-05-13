package routes

import (
	"github.com/chocological13/kittykeeper/services/user-service/internal/handlers"
	"github.com/chocological13/kittykeeper/services/user-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetUpRoutes(rg *gin.RouterGroup, userHandler *handlers.UserHandler, authMiddleware *middleware.AuthMiddleware) {
	rg.POST("/register", userHandler.Register)
	rg.POST("/login", userHandler.Login)
	rg.POST("/refresh", userHandler.RefreshToken)

	authRoutes := rg.Group("/")
	authRoutes.Use(authMiddleware.RequireAuth())
	{
		authRoutes.GET("/profile", userHandler.GetProfile)
		authRoutes.POST("/logout", userHandler.Logout)
	}
}
