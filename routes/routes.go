package routes

import (
	"urllite/handler"

	"github.com/gin-gonic/gin"
)

func MountHTTPRoutes(r *gin.Engine) {
	userHandlers := handler.NewUserHandler()
	userGroup := r.Group("/api/v1/user")
	{
		userGroup.POST("/", userHandlers.CreateUser)
	}
}

