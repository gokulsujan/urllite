package main

import (
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
	routes.MountHTTPRoutes(r)

	r.Run()
}
