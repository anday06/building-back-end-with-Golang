package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	JWTSecret      string
	JWTExpiresHour int
}

func Load() Config {
	hours, _ := strconv.Atoi(getEnv("JWT_EXPIRES_HOURS", "24"))
	return Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=tasks port=5432 sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:      getEnv("JWT_SECRET", "development-secret"),
		JWTExpiresHour: hours,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
