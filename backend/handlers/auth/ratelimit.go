package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/yourusername/kor-assetforge/backend/apperrors"
)

// AuthRateLimiter handles rate limiting for authentication endpoints
type AuthRateLimiter struct {
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

// NewAuthRateLimiter creates a new auth rate limiter
func NewAuthRateLimiter(r rate.Limit, b int) *AuthRateLimiter {
	return &AuthRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getLimiter gets or creates a rate limiter for the given key
func (rl *AuthRateLimiter) getLimiter(key string) *rate.Limiter {
	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}
	return limiter
}

// LoginRateLimit middleware limits login attempts per IP
func (rl *AuthRateLimiter) LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "login:" + c.ClientIP()
		limiter := rl.getLimiter(key)

		if !limiter.Allow() {
			apperrors.AbortWithError(c, apperrors.NewTooManyRequestsError("Too many login attempts. Please try again later."))
			return
		}

		c.Next()
	}
}

// RegisterRateLimit middleware limits registration attempts per IP
func (rl *AuthRateLimiter) RegisterRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "register:" + c.ClientIP()
		limiter := rl.getLimiter(key)

		if !limiter.Allow() {
			apperrors.AbortWithError(c, apperrors.NewTooManyRequestsError("Too many registration attempts. Please try again later."))
			return
		}

		c.Next()
	}
}

// PasswordResetRateLimit middleware limits password reset attempts per IP
func (rl *AuthRateLimiter) PasswordResetRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "password_reset:" + c.ClientIP()
		limiter := rl.getLimiter(key)

		if !limiter.Allow() {
			apperrors.AbortWithError(c, apperrors.NewTooManyRequestsError("Too many password reset attempts. Please try again later."))
			return
		}

		c.Next()
	}
}

// EmailVerificationRateLimit middleware limits email verification attempts per IP
func (rl *AuthRateLimiter) EmailVerificationRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "email_verify:" + c.ClientIP()
		limiter := rl.getLimiter(key)

		if !limiter.Allow() {
			apperrors.AbortWithError(c, apperrors.NewTooManyRequestsError("Too many email verification attempts. Please try again later."))
			return
		}

		c.Next()
	}
}

// GeneralAuthRateLimit provides general rate limiting for auth endpoints
func (rl *AuthRateLimiter) GeneralAuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "auth:" + c.ClientIP()
		limiter := rl.getLimiter(key)

		if !limiter.Allow() {
			apperrors.AbortWithError(c, apperrors.NewTooManyRequestsError("Too many authentication requests. Please try again later."))
			return
		}

		c.Next()
	}
}