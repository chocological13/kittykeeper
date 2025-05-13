package main

import (
	"context"
	"errors"
	"github.com/chocological13/kittykeeper/services/user-service/internal/auth"
	"github.com/chocological13/kittykeeper/services/user-service/internal/database/repository"
	"github.com/chocological13/kittykeeper/services/user-service/internal/handlers"
	"github.com/chocological13/kittykeeper/services/user-service/internal/middleware"
	"github.com/chocological13/kittykeeper/services/user-service/internal/routes"
	"github.com/chocological13/kittykeeper/services/user-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// StartServer starts the user service server
func StartServer(r *gin.Engine, port string, db *pgxpool.Pool, redisClient *redis.Client,
	authCfg *auth.AuthConfig) {
	// ! CORS middleware
	r.Use(middleware.SetupCORS())

	queries := repository.New(db)

	authService := auth.NewAuthService(authCfg)
	tokenStore := auth.NewTokenStore(redisClient, authCfg.AccessTokenTTL, authCfg.RefreshTokenTTL)
	userService := service.NewUserService(queries, authService, tokenStore)
	authMiddleware := middleware.NewAuthMiddleware(authService, tokenStore)

	userHandler := handlers.NewUserHandler(userService)

	v1 := r.Group("/api/v1")
	routes.SetUpRoutes(v1, userHandler, authMiddleware)

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Listening on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start server: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown server: %s\n", err)
	}

	log.Info("Server exited properly")
}
