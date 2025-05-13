package database

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

func ConnectDB(connString string) *pgxpool.Pool {
	dbPool, err := pgxpool.New(context.Background(), connString)
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
		log.Fatalf("failed to get current working directory: %v", err)
		return err
	}
	migrationsDir := filepath.Join(dir, "migrations")

	return goose.Up(db, migrationsDir)
}
