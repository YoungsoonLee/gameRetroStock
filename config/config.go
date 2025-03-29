package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	IGDBClientID     string
	IGDBClientSecret string
	DatabaseURL      string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	return &Config{
		Port:             getEnv("PORT", "8080"),
		IGDBClientID:     getEnv("IGDB_CLIENT_ID", ""),
		IGDBClientSecret: getEnv("IGDB_CLIENT_SECRET", ""),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
