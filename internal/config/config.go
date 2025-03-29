package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	IGDBClientID     string
	IGDBClientSecret string
	EbayAppID        string
	EbayDevID        string
	EbayCertID       string
	DatabaseURL      string
}

var AppConfig Config

func Init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Load configuration from environment variables
	AppConfig = Config{
		IGDBClientID:     getEnvOrDefault("IGDB_CLIENT_ID", ""),
		IGDBClientSecret: getEnvOrDefault("IGDB_CLIENT_SECRET", ""),
		EbayAppID:        getEnvOrDefault("EBAY_APP_ID", ""),
		EbayDevID:        getEnvOrDefault("EBAY_DEV_ID", ""),
		EbayCertID:       getEnvOrDefault("EBAY_CERT_ID", ""),
		DatabaseURL:      constructDatabaseURL(),
	}

	// Validate required configuration
	if AppConfig.IGDBClientID == "" || AppConfig.IGDBClientSecret == "" {
		log.Fatal("IGDB credentials are required")
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func constructDatabaseURL() string {
	// First try to get the complete DATABASE_URL
	if dbURL := os.Getenv("DB_URL"); dbURL != "" {
		return dbURL
	}

	// If not available, construct from individual components
	dbUser := getEnvOrDefault("DB_USER", "admin")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "test1234")
	dbHost := getEnvOrDefault("DB_HOST", "127.0.0.1")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbName := getEnvOrDefault("DB_NAME", "gameretrostock")

	return "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Get database URL or construct it
	databaseURL := os.Getenv("DATABASE_URL")
	log.Println("databaseURL:", databaseURL)
	if databaseURL == "" {
		// Use default connection string
		databaseURL = constructDatabaseURL()
	}

	// Get IGDB credentials
	igdbClientID := os.Getenv("IGDB_CLIENT_ID")
	if igdbClientID == "" {
		igdbClientID = "default_client_id" // Use default for development
	}

	igdbClientSecret := os.Getenv("IGDB_CLIENT_SECRET")
	if igdbClientSecret == "" {
		igdbClientSecret = "default_client_secret" // Use default for development
	}

	cfg := &Config{
		DatabaseURL:      databaseURL,
		IGDBClientID:     igdbClientID,
		IGDBClientSecret: igdbClientSecret,
	}

	return cfg, nil
}
