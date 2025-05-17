package main

import (
	"context"
	"errors"
	"github.com/chocological13/kittykeeper/cat-service/internal/auth"
	"github.com/chocological13/kittykeeper/cat-service/internal/config"
	"github.com/chocological13/kittykeeper/cat-service/internal/database/repository"
	"github.com/chocological13/kittykeeper/cat-service/internal/handlers"
	"github.com/chocological13/kittykeeper/cat-service/internal/middleware"
	"github.com/chocological13/kittykeeper/cat-service/internal/routes"
	"github.com/chocological13/kittykeeper/cat-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartServer(db *pgxpool.Pool, rdb *redis.Client, cfg *config.Config) {
	r := gin.Default()
	r.Use(middleware.SetupCORS())

	// TODO : set up service and handlers
	queries := repository.New(db)
	catService := service.NewCatService(queries)
	jwtService := auth.NewJWTService(cfg.AccessTokenSecret)
	tokenStore := auth.NewTokenStore(rdb)

	authMiddleware := middleware.NewAuthMiddleware(jwtService, tokenStore, log)
	catHandler := handlers.NewCatHandler(catService, log)

	// TODO : set up routes
	rg := r.Group("/api/v1")
	routes.SetUpRoutes(rg, authMiddleware, catHandler)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Listening on port %s\n", cfg.Port)
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
