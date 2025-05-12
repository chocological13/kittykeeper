package main

import (
	"context"
	"errors"
	"github.com/chocological13/kittykeeper/services/user-service/config"
	"github.com/chocological13/kittykeeper/services/user-service/internal/auth"
	"github.com/chocological13/kittykeeper/services/user-service/internal/database"
	"github.com/chocological13/kittykeeper/services/user-service/internal/database/repository"
	"github.com/chocological13/kittykeeper/services/user-service/internal/handlers"
	"github.com/chocological13/kittykeeper/services/user-service/internal/logger"
	"github.com/chocological13/kittykeeper/services/user-service/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var log = logger.NewLogger("user-service")

func main() {
	log.Info("Starting user service")

	// Load config
	cfg, authCfg, err := config.LoadConfig()
	if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}

	// Connect to database
	db := database.ConnectDB(cfg.DatabaseUrl)
	log.Info("Connected to database")
	if db == nil {
		log.WithError(err).Fatal("failed to connect to database")
	}
	defer db.Close()

	// Connect to redis
	redisClient := database.ConnectRedis(cfg.RedisUrl)
	log.Info("Connected to user redis")
	if redisClient == nil {
		log.WithError(err).Fatal("failed to connect to redis")
	}
	defer redisClient.Close()

	// Run migrations
	err = database.RunMigrations(cfg.DatabaseUrl)
	if err != nil {
		log.WithError(err).Fatal("failed to run migrations")
	}
	log.Info("Migrations ran successfully")

	// Setup gin router
	r := gin.Default()

	// TODO : services and handlers
	queries := repository.New(db)

	authService := auth.NewAuthService(authCfg)
	tokenStore := auth.NewTokenStore(redisClient, authCfg.AccessTokenTTL, authCfg.RefreshTokenTTL)
	userService := service.NewUserService(queries, authService, tokenStore)

	userHandler := handlers.NewUserHandler(userService)

	// TODO : add routes / create route set up separately
	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Start server
	startServer(r, cfg.Port)
}

// startServer is a helper that starts the user service server
func startServer(r *gin.Engine, port string) {
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
