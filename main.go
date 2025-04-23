package main

import (
	"os"
	"time"
	"urllite/config/database"
	"urllite/config/env"
	"urllite/routes"
	"urllite/store"

	"github.com/gin-contrib/cors"
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
	r.RedirectTrailingSlash = false
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(204)
	})
	routes.MountHTTPRoutes(r)

	r.Run()
}
