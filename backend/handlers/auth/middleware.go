package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/yourusername/kor-assetforge/backend/apperrors"
	"github.com/yourusername/kor-assetforge/backend/models"
)

// AuthMiddleware handles JWT authentication and authorization
type AuthMiddleware struct {
	jwtSecret string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

// JWTAuth middleware validates JWT tokens
func (m *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apperrors.AbortWithError(c, apperrors.NewUnauthorizedError("Authorization header required"))
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			apperrors.AbortWithError(c, apperrors.NewUnauthorizedError("Invalid authorization header format"))
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperrors.NewUnauthorizedError("Invalid signing method")
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil {
			apperrors.AbortWithError(c, apperrors.NewUnauthorizedError("Invalid token"))
			return
		}

		if !token.Valid {
			apperrors.AbortWithError(c, apperrors.NewUnauthorizedError("Invalid token"))
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			apperrors.AbortWithError(c, apperrors.NewUnauthorizedError("Invalid token claims"))
			return
		}

		// Check token type
		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != "access" {
			apperrors.AbortWithError(c, apperrors.NewUnauthorizedError("Invalid token type"))
			return
		}

		// Extract user information
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			apperrors.AbortWithError(c, apperrors.NewUnauthorizedError("Invalid user ID in token"))
			return
		}
		userID := uint(userIDFloat)

		email, _ := claims["email"].(string)
		username, _ := claims["username"].(string)
		role, _ := claims["role"].(string)

		// Set user information in context
		c.Set("user_id", userID)
		c.Set("email", email)
		c.Set("username", username)
		c.Set("role", role)

		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func (m *AuthMiddleware) RequireRole(requiredRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			apperrors.AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
			return
		}

		userRole := models.UserRole(role.(string))

		// Check role hierarchy: admin > moderator > user
		if !hasRequiredRole(userRole, requiredRole) {
			apperrors.AbortWithError(c, apperrors.NewForbiddenError("Insufficient permissions"))
			return
		}

		c.Next()
	}
}

// RequireRoles middleware checks if user has any of the required roles
func (m *AuthMiddleware) RequireRoles(requiredRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			apperrors.AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
			return
		}

		userRole := models.UserRole(role.(string))

		hasRole := false
		for _, requiredRole := range requiredRoles {
			if hasRequiredRole(userRole, requiredRole) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			apperrors.AbortWithError(c, apperrors.NewForbiddenError("Insufficient permissions"))
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware sets user context if token is present but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, nil
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		// Check token type
		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != "access" {
			c.Next()
			return
		}

		// Extract user information
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			c.Next()
			return
		}
		userID := uint(userIDFloat)

		email, _ := claims["email"].(string)
		username, _ := claims["username"].(string)
		role, _ := claims["role"].(string)

		// Set user information in context
		c.Set("user_id", userID)
		c.Set("email", email)
		c.Set("username", username)
		c.Set("role", role)
		c.Set("authenticated", true)

		c.Next()
	}
}

// hasRequiredRole checks if user role meets the required role level
func hasRequiredRole(userRole, requiredRole models.UserRole) bool {
	roleHierarchy := map[models.UserRole]int{
		models.RoleUser:      1,
		models.RoleModerator: 2,
		models.RoleAdmin:     3,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}