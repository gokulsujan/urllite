package routes

import (
	"urllite/auth"
	"urllite/handler"

	"github.com/gin-gonic/gin"
)

func MountHTTPRoutes(r *gin.Engine) {
	userHandlers := handler.NewUserHandler()
	r.POST("/signup", userHandlers.Signup)
	r.POST("/login", userHandlers.Login)
	r.POST("/change-password", auth.UserAuthentication, userHandlers.ChangePassword)

	userGroup := r.Group("/api/v1/user")
	{
		userGroup.POST("/", auth.UserAuthentication, userHandlers.CreateUser)
		userGroup.GET("/", auth.UserAuthentication, userHandlers.GetUsers)
		userGroup.GET("/:id", auth.UserAuthentication, userHandlers.GetUserByID)
		userGroup.PATCH("/:id", auth.UserAuthentication, userHandlers.UpdateUserByID)
		userGroup.DELETE("/:id", auth.UserAuthentication, userHandlers.DeleteUserByID)
	}
}
