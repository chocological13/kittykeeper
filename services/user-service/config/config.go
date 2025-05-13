package config

import (
	"github.com/chocological13/kittykeeper/services/user-service/internal/auth"
	"github.com/chocological13/kittykeeper/services/user-service/internal/logger"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseUrl string
	Port        string
	Environment string
	RedisUrl    string
}

var log = logger.NewLogger("user-service")

// Loads config from environment variables
func LoadConfig() (*Config, *auth.AuthConfig, error) {
	var err error
	if os.Getenv("ENVIRONMENT") != "production" {
		err = loadGoDotEnv()
	}
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := os.Getenv("USER_DB_URL")
	port := os.Getenv("USER_PORT")
	environment := os.Getenv("ENVIRONMENT")
	redisAddr := os.Getenv("USER_REDIS_ADDR")

	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	refreshTokenSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	accessTokenTTLString := os.Getenv("ACCESS_TOKEN_TTL")

	accessTokenTTL, err := strconv.Atoi(accessTokenTTLString)
	if err != nil {
		log.Fatalf("Error parsing ACCESS_TOKEN_TTL: %v", err)
	}

	refreshTokenTTLString := os.Getenv("REFRESH_TOKEN_TTL")
	refreshTokenTTL, err := strconv.Atoi(refreshTokenTTLString)
	if err != nil {
		log.Fatalf("Error parsing REFRESH_TOKEN_TTL: %v", err)
	}

	config := &Config{
		DatabaseUrl: dsn,
		Port:        port,
		Environment: environment,
		RedisUrl:    redisAddr,
	}

	securityConfig := &auth.AuthConfig{
		AccessTokenSecret:  accessTokenSecret,
		RefreshTokenSecret: refreshTokenSecret,
		AccessTokenTTL:     time.Second * time.Duration(accessTokenTTL),
		RefreshTokenTTL:    time.Second * time.Duration(refreshTokenTTL),
	}

	return config, securityConfig, nil
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
