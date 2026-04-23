package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"

	"github.com/yourusername/kor-assetforge/backend/apperrors"
	"github.com/yourusername/kor-assetforge/backend/config"
	"github.com/yourusername/kor-assetforge/backend/handlers"
	"github.com/yourusername/kor-assetforge/backend/handlers/auth"
	"github.com/yourusername/kor-assetforge/backend/models"
	"github.com/yourusername/kor-assetforge/backend/utils"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Stellar client
	stellarClient, err := config.InitStellarClient()
	if err != nil {
		log.Fatalf("Failed to initialize Stellar client: %v", err)
	}

	// Initialize Redis
	redisURL := os.Getenv("REDIS_URL")
	redisClient, err := utils.InitRedis(redisURL)
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis, continuing without cache: %v", err)
		redisClient = nil
	} else {
		defer redisClient.Close()
	}

	// Initialize cache
	var cache *utils.Cache
	if redisClient != nil {
		cache = utils.NewCache(redisClient)
	}

	// Setup authentication configuration
	authConfig := &auth.AuthConfig{
		JWTSecret:           getEnvOrDefault("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpirationHours:  getEnvIntOrDefault("JWT_EXPIRATION_HOURS", 24),
		RefreshTokenHours:   getEnvIntOrDefault("REFRESH_TOKEN_HOURS", 168), // 7 days
		EmailTokenHours:     getEnvIntOrDefault("EMAIL_TOKEN_HOURS", 24),
		PasswordResetHours:  getEnvIntOrDefault("PASSWORD_RESET_HOURS", 1),
		BcryptCost:          getEnvIntOrDefault("BCRYPT_COST", 12),
	}

	// Initialize auth components
	authHandler := auth.NewAuthHandler(db, authConfig, cache)
	authMiddleware := auth.NewAuthMiddleware(authConfig.JWTSecret)

	// Setup rate limiter for auth endpoints (5 requests per minute, burst of 10)
	authRateLimiter := auth.NewAuthRateLimiter(rate.Limit(5.0/60.0), 10)

	// Setup router
	router := gin.New() // Use gin.New() instead of gin.Default() to avoid default logger/recovery

	debugMode := strings.EqualFold(os.Getenv("DEBUG_MODE"), "true")

	// Use custom enhanced middleware
	router.Use(
		handlers.RequestLogger(),
		apperrors.ErrorHandler(debugMode),
	)

	router.GET("/metrics", apperrors.MetricsHandler)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "kor-AssetForge API",
			"version": "0.1.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		authGroup := v1.Group("/auth")
		authGroup.Use(authRateLimiter.GeneralAuthRateLimit())
		{
			authGroup.POST("/register", authRateLimiter.RegisterRateLimit(), authHandler.Register)
			authGroup.POST("/login", authRateLimiter.LoginRateLimit(), authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.POST("/verify-email", authRateLimiter.EmailVerificationRateLimit(), authHandler.VerifyEmail)
			authGroup.POST("/forgot-password", authRateLimiter.PasswordResetRateLimit(), authHandler.ForgotPassword)
			authGroup.POST("/reset-password", authHandler.ResetPassword)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(authMiddleware.JWTAuth())
		{
			// User profile
			protected.GET("/profile", authHandler.GetProfile)
			protected.POST("/logout", authHandler.Logout)

			// Asset routes (require authentication)
			assetHandler := handlers.NewAssetHandler(db, stellarClient, redisClient)
			protected.POST("/assets/tokenize", assetHandler.TokenizeAsset)
			protected.POST("/assets", assetHandler.TokenizeAsset)
			protected.GET("/assets", assetHandler.ListAssets)
			protected.GET("/assets/:id", assetHandler.GetAsset)

			// Marketplace routes
			protected.POST("/marketplace/list", assetHandler.ListAssetForSale)
			protected.POST("/marketplace/transfer", assetHandler.TransferAsset)
			protected.GET("/transactions", assetHandler.ListTransactions)

			// Admin only routes
			admin := protected.Group("")
			admin.Use(authMiddleware.RequireRole(models.RoleAdmin))
			{
				// Admin-specific routes can be added here
			}
		}

		// Webhook routes (public but with HMAC verification)
		webhookHandler := handlers.NewWebhookHandler(db)
		router.POST("/webhooks/stellar-events", webhookHandler.HandleStellarEvent)
	}

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault returns environment variable value as int or default
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
