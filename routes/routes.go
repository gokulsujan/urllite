package routes

import (
	"urllite/auth"
	"urllite/handler"
	admin_handlers "urllite/handler/admin"
	"urllite/security"

	"github.com/gin-gonic/gin"
)

func MountHTTPRoutes(r *gin.Engine) {
	userHandlers := handler.NewUserHandler()
	urlHandler := handler.NewUrlHandler()
	adminhandlers := admin_handlers.NewAdminAuthHandler()
	adminUserHandler := admin_handlers.NewAdminUserHandler()
	adminUrlHandler := admin_handlers.NewAdminUrlHandler()
	r.POST("/signup", security.RatelimittingMiddleware, userHandlers.Signup)
	r.POST("/signup-and-login", security.RatelimittingMiddleware, userHandlers.SignupAndLogin)
	r.POST("/login", security.RatelimittingMiddleware, userHandlers.Login)
	r.POST("/send-forget-password-otp", security.OtpRatelimittingMiddleware, userHandlers.SendForgetPasswordOtp)
	r.POST("/verify-forget-password-otp", userHandlers.VerifyForgetPasswordOtp)
	r.POST("/change-password-via-otp", userHandlers.ChangePasswordUsingOtp)
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

		r.POST("/api/v1/admin/login", adminhandlers.Login)
		adminGroup := authenticatedApis.Group("/admin", auth.AdminAuthentication)
		{
			adminGroup.GET("/dashboard", adminhandlers.Dashboard)
			adminGroup.GET("/user/:id/stats", adminUserHandler.UserDashboardStats)
			adminGroup.GET("/user/:id/usage", adminUserHandler.UserUsageStats)
			adminGroup.GET("/user/:id/urls", adminUrlHandler.UrlsByUserID)
		}
	}
}
