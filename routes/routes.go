package routes

import (
	"urllite/handler"

	"github.com/gin-gonic/gin"
)

func MountHTTPRoutes(r *gin.Engine) {
	userHandlers := handler.NewUserHandler()
	r.POST("/signup", userHandlers.Signup)
	r.POST("/login", userHandlers.Login)

	userGroup := r.Group("/api/v1/user")
	{
		userGroup.POST("/", userHandlers.CreateUser)
		userGroup.GET("/", userHandlers.GetUsers)
		userGroup.GET("/:id", userHandlers.GetUserByID)
		userGroup.PATCH("/:id", userHandlers.UpdateUserByID)
		userGroup.DELETE("/:id", userHandlers.DeleteUserByID)
	}
}
