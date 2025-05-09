package main

import (
	"context"
	"errors"
	"github.com/chocological13/kittykeeper/services/user-service/config"
	"github.com/chocological13/kittykeeper/services/user-service/internal/database"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config for user service: %v", err)
	}

	// Connect to database
	dbPool := database.ConnectDB(cfg.DatabaseUrl)
	log.Println("Connected to user service database")
	defer dbPool.Close()

	// Connect to redis
	redisClient := database.ConnectRedis(cfg.RedisUrl)
	log.Println("Connected to user service redis")
	defer redisClient.Close()

	// Run migrations
	err = database.RunMigrations(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("user service failed to run migrations: %v", err)
	}
	log.Println("User service migrations ran successfully")

	// Setup gin router
	r := gin.Default()

	// TODO : add routes / create route set up separately

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Printf("User service listening on port %s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start user service server: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down user service server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown user service server: %s\n", err)
	}

	log.Println("User service server exited properly")
}
