package main

import (
	"urllite/config/database"
	"urllite/config/env"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Load environment variables
	env.EnableEnvVariables()
	// Connect to the database
	database.Connect()

	r.Run()
}
