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
	r.POST("/signup-and-login", security.RatelimittingMiddleware, userHandlers.SignupAndLogin)
	r.POST("/login", security.RatelimittingMiddleware, userHandlers.Login)
	r.POST("/change-password", auth.UserAuthentication, userHandlers.ChangePassword)
	r.POST("/verify-email-otp", security.OtpRatelimittingMiddleware, userHandlers.SendEmailVerificationOtp)
	r.POST("/verify-email", userHandlers.VerifyEmail)
	r.GET("/:short_url", urlHandler.RedirectToLongUrl)

	authenticatedApis := r.Group("/api/v1", auth.UserAuthentication)
	{
		authenticatedApis.GET("/profile", userHandlers.Profile)
		userGroup := authenticatedApis.Group("/user")
		{
			userGroup.POST("/", userHandlers.CreateUser)
			userGroup.GET("/:id", userHandlers.GetUserByID)
			userGroup.PUT("/:id", userHandlers.UpdateUserByID)

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
			urlGroup.GET("/:id/logs", urlHandler.GetUrlLogsByUrl)

		}
	}
}
