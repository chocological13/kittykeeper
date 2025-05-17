package routes

import (
	"github.com/chocological13/kittykeeper/cat-service/internal/handlers"
	"github.com/chocological13/kittykeeper/cat-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetUpRoutes(rg *gin.RouterGroup, middleware *middleware.AuthMiddleware, catHandler *handlers.CatHandler) {
	rg.Use(middleware.RequireAuth())
	{
		rg.POST("/cats", catHandler.CreateCat)
	}
}
