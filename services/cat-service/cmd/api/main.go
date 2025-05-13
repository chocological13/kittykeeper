package main

import (
	"github.com/chocological13/kittykeeper/cat-service/internal/config"
	"github.com/chocological13/kittykeeper/cat-service/internal/database"
	"github.com/chocological13/kittykeeper/cat-service/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var log = logger.NewLogger("cat-service")

func main() {
	log.Info("Starting cat service")

	// * Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}

	// * Connect to database
	db := database.ConnectDB(cfg.DatabaseUrl)
	if db == nil {
		log.WithError(err).Fatal("failed to connect to database")
	}
	log.Info("Connected to database")
	defer db.Close()

	// * Run migrations
	err = database.RunMigrations(cfg.DatabaseUrl)
	if err != nil {
		log.WithError(err).Fatal("failed to run migrations")
	}
	log.Info("Migrations ran successfully")

	// * Start server
	StartServer(db, cfg)
}
