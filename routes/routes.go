package routes

import (
	"urllite/auth"
	"urllite/handler"

	"github.com/gin-gonic/gin"
)

func MountHTTPRoutes(r *gin.Engine) {
	userHandlers := handler.NewUserHandler()
	urlHandler := handler.NewUrlHandler()
	r.POST("/signup", userHandlers.Signup)
	r.POST("/login", userHandlers.Login)
	r.POST("/change-password", auth.UserAuthentication, userHandlers.ChangePassword)
	r.GET("/:short_url", urlHandler.RedirectToLongUrl)

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
			urlGroup.POST("/", urlHandler.Create)
			urlGroup.GET("/", urlHandler.GetURLs)
			urlGroup.GET("/:id", urlHandler.GetUrlByID)
			urlGroup.DELETE("/:id", urlHandler.DeleteURLById)
		}
	}
}
