package main

import (
	"os"
	"urllite/config/database"
	"urllite/config/env"
	"urllite/routes"
	"urllite/store"

	"github.com/gin-gonic/gin"
)

func init() {
	env.EnableEnvVariables()
	database.Connect()
	store.AutoMigrateTables()
}

func main() {
	r := gin.Default()
	if os.Getenv("production") == "true" {
		gin.SetMode(gin.ReleaseMode)
	}
	routes.BypassCorsPolicy(r)
	routes.MountHTTPRoutes(r)

	r.Run()
}
