package config

import (
	"github.com/chocological13/kittykeeper/services/user-service/internal/auth"
	"github.com/joho/godotenv"
	"log"
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

// Loads config from environment variables
func LoadConfig() (*Config, *auth.AuthConfig, error) {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	environment := os.Getenv("ENV")
	redisAddr := os.Getenv("REDIS_ADDR")

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
