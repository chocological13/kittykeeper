package main

import (
	"github.com/chocological13/kittykeeper/services/user-service/config"
	"github.com/chocological13/kittykeeper/services/user-service/internal/database"
	"github.com/chocological13/kittykeeper/services/user-service/internal/logger"
	"github.com/gin-gonic/gin"
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

	// Start server
	StartServer(r, cfg.Port, db, redisClient, authCfg)
}
