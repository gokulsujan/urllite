package security

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RatelimittingMiddleware(c *gin.Context) {
	rl := NewRateLimitter(c)
	limitter := rl.GetLimitter()
	if !limitter.Allow() {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"status": "failed", "message": "Request limit exceeded. Wait for 5 seconds"})
		return
	}

}

func OtpRatelimittingMiddleware(c *gin.Context) {
	rl := NewRateLimitter(c)
	limitter := rl.GetOtpLimitter()
	if !limitter.Allow() {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"status": "failed", "message": "Otp limit exceeded. Retry after 2 minutes"})
		return
	}

}
