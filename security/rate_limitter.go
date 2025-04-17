package security

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	limiters  = make(map[string]*rate.Limiter)
	mu        sync.Mutex
	rateLimit = rate.Every(5 * time.Second)
	burst     = 5
)

type RateLimitter interface {
	GetLimitter() *rate.Limiter
}

type rateLimitter struct {
	context     *gin.Context
}

func NewRateLimitter(c *gin.Context) RateLimitter {
	return &rateLimitter{context: c}
}

func (rl *rateLimitter) GetLimitter() *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	clientIP := rl.context.ClientIP()
	limitter, exists := limiters[clientIP]
	if !exists {
		limitter = rate.NewLimiter(rateLimit, burst)
		limiters[clientIP] = limitter
	}

	return limitter
}
