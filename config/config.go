package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Env func(string) string

// will run automatically when package is imported
func init() {
	// Load .env from config/.env
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	Env = os.Getenv
}
