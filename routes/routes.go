package routes

import (
	"encoding/json"
	"log"
	"os"
	"time"
	"urllite/auth"
	"urllite/handler"
	admin_handlers "urllite/handler/admin"
	"urllite/security"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func MountHTTPRoutes(r *gin.Engine) {
	r.LoadHTMLGlob("utils/assets/html/*.html")
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

func BypassCorsPolicy(r *gin.Engine) {
	raw := os.Getenv("WHITELISTED_APIS")
	var whitelist []string
	err := json.Unmarshal([]byte(raw), &whitelist)
	if err != nil {
		log.Fatalf("Failed to parse WHITELISTED_APIS: %v", err)
	}

	r.RedirectTrailingSlash = false
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://admin.urllite.in", "https://app.urllite.in", "http://localhost:5173", "http://localhost:3000", "http://192.168.1.4:3000", "http://localhost:3001", "http://192.168.1.4:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(204)
	})
}
