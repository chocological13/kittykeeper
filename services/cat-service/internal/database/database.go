package database

import (
	"context"
	"github.com/chocological13/kittykeeper/cat-service/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"os"
	"path/filepath"
)

var log = logger.NewLogger("cat-service")

func ConnectDB(dbUrl string) *pgxpool.Pool {
	dbPool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		panic(err)
	}
	return dbPool
}

func ConnectRedis(redisAddr string) *redis.Client {
	opt, err := redis.ParseURL(redisAddr)
	if err != nil {
		panic(err)
	}
	return redis.NewClient(opt)
}

func RunMigrations(dbUrl string) error {
	db, err := goose.OpenDBWithDriver("postgres", dbUrl)
	if err != nil {
		return err
	}
	defer db.Close()

	dir, err := os.Getwd()
	if err != nil {
		log.WithError(err).Fatal("failed to get current working directory")
		return err
	}
	migrationsDir := filepath.Join(dir, "internal/database/migrations")
	return goose.Up(db, migrationsDir)
}
