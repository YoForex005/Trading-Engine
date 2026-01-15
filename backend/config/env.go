package config

import (
	"log"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file from project root before any other package initialization
	// This must run before auth package's init() which reads JWT_SECRET
	// Try multiple paths to handle different working directories
	if err := godotenv.Load("../.env"); err != nil {
		if err2 := godotenv.Load(".env"); err2 != nil {
			// Both failed - env vars must be set directly (production mode)
			log.Printf("[ENV] .env file not found, using system environment variables")
		}
	}
}
