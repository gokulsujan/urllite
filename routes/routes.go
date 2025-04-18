package routes

import (
	"urllite/auth"
	"urllite/handler"
	"urllite/security"

	"github.com/gin-gonic/gin"
)

func MountHTTPRoutes(r *gin.Engine) {
	userHandlers := handler.NewUserHandler()
	urlHandler := handler.NewUrlHandler()
	r.POST("/signup", security.RatelimittingMiddleware, userHandlers.Signup)
	r.POST("/login", security.RatelimittingMiddleware, userHandlers.Login)
	r.POST("/change-password", auth.UserAuthentication, userHandlers.ChangePassword)
	r.POST("/verify-email-otp", userHandlers.SendEmailVerificationOtp)
	r.POST("/verify-email", userHandlers.VerifyEmail)
	r.GET("/:short_url", urlHandler.RedirectToLongUrl)

	authenticatedApis := r.Group("/api/v1", auth.UserAuthentication)
	{
		userGroup := authenticatedApis.Group("/user")
		{
			userGroup.POST("/", userHandlers.CreateUser)
			userGroup.GET("/:id", userHandlers.GetUserByID)
			userGroup.PATCH("/:id", userHandlers.UpdateUserByID)

			userGroup.GET("/", auth.AdminAuthentication, userHandlers.GetUsers)
			userGroup.DELETE("/:id", auth.AdminAuthentication, userHandlers.DeleteUserByID)
			userGroup.POST("/:id/make-admin", auth.AdminAuthentication, userHandlers.MakeAdmin)
		}

		urlGroup := authenticatedApis.Group("/url")
		{
			urlGroup.POST("/", security.RatelimittingMiddleware, urlHandler.Create)
			urlGroup.GET("/", urlHandler.GetURLs)
			urlGroup.GET("/:id", urlHandler.GetUrlByID)
			urlGroup.DELETE("/:id", urlHandler.DeleteURLById)
		}
	}
}
