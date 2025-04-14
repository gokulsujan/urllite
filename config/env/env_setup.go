package env

import (
	"log"

	"github.com/joho/godotenv"
)

func EnableEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
