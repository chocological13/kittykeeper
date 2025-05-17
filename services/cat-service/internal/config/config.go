package config

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	DatabaseUrl string
	Port        string
	Environment string
	RedisAddr   string

	AccessTokenSecret string
}

func LoadConfig() (*Config, error) {
	var err error
	if os.Getenv("ENVIRONMENT") != "production" {
		err = loadGoDotEnv()
	}
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := os.Getenv("CAT_DB_URL")
	port := os.Getenv("CAT_PORT")
	environment := os.Getenv("ENVIRONMENT")
	redisAddr := os.Getenv("REDIS_ADDR")

	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")

	config := &Config{
		DatabaseUrl: dsn,
		Port:        port,
		Environment: environment,
		RedisAddr:   redisAddr,

		AccessTokenSecret: accessTokenSecret,
	}

	return config, nil
}

func loadGoDotEnv() error {
	var err error

	// Load .env file from current directory, parent directory, parent of parent directory, and finally /app/.env
	paths := []string{
		".env",
		"../.env",
		"../../.env",
		"/app/.env", // Docker context
	}

	for _, path := range paths {
		if err = godotenv.Load(path); err == nil {
			log.Infof("Loaded env file from %s", path)
			break
		}
	}

	if err != nil {
		log.Warn("No .env file found for local development")
		return err
	}

	log.Info("Loaded .env file for local development")
	return nil
}
