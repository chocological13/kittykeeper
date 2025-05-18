package routes

import (
	"github.com/chocological13/kittykeeper/cat-service/internal/handlers"
	"github.com/chocological13/kittykeeper/cat-service/internal/middleware"
	"github.com/chocological13/kittykeeper/cat-service/internal/service"
	"github.com/gin-gonic/gin"
)

func SetUpRoutes(rg *gin.RouterGroup, middleware *middleware.AuthMiddleware,
	catService *service.CatService, catHandler *handlers.CatHandler) {

	cats := rg.Group("/cats", middleware.RequireAuth())
	{
		// ? Endpoints that do not require ownership check
		cats.POST("", catHandler.CreateCat)
		cats.GET("", catHandler.ListCatsByOwner)

		// ? Endpoints that require ownership check
		catResource := cats.Group("/:id")
		{
			catResource.Use(middleware.OwnershipCheck(catService))

			catResource.GET("", catHandler.GetCat)
			catResource.PUT("", catHandler.UpdateCat)
			catResource.PATCH("/clear-death", catHandler.ClearDateOfDeath)
		}
	}
}
