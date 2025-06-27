package initializers

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() {
	// Only load .env file in non-production environments
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: .env file not found, continuing without it")
		}
	}

	// Fail early if critical environment variable is missing
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is not set in the environment")
	}
}
