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

	authenticatedApis := r.Group("/api/v1", auth.UserAuthentication)
	{
		userGroup := authenticatedApis.Group("/user")
		{
			userGroup.POST("/", userHandlers.CreateUser)
			userGroup.GET("/", userHandlers.GetUsers)
			userGroup.GET("/:id", userHandlers.GetUserByID)
			userGroup.PATCH("/:id", userHandlers.UpdateUserByID)
			userGroup.DELETE("/:id", userHandlers.DeleteUserByID)
		}

		urlGroup := authenticatedApis.Group("/url")
		{
			urlGroup.POST("/")
		}
	}
}
