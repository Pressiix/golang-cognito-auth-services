package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads the .env file and returns the value for a given key
func LoadEnv(key string) string {
	// Load .env file (use .env or .env.local)
	if err := godotenv.Load(".env"); err != nil {
		// If .env.local exists, try loading that
		if err := godotenv.Load(".env.local"); err != nil {
			log.Printf("Warning: No .env file found")
		}
	}

	return os.Getenv(key)
}

// MustLoadEnv loads the .env file and fatally exits if the key is not found
func MustLoadEnv(key string) string {
	value := LoadEnv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return value
}