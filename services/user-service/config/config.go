package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	DatabaseUrl string
	Port        string
	JwtSecret   string
	Environment string
	RedisUrl    string
}

// Loads config from environment variables
func LoadConfig() (*Config, error) {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	environment := os.Getenv("ENV")
	redisAddr := os.Getenv("REDIS_ADDR")

	return &Config{
		DatabaseUrl: dsn,
		Port:        port,
		JwtSecret:   jwtSecret,
		Environment: environment,
		RedisUrl:    redisAddr,
	}, nil
}
