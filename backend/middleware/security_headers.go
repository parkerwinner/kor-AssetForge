package middleware

import "github.com/gin-gonic/gin"

const (
	contentSecurityPolicy   = "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; object-src 'none'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'"
	strictTransportSecurity = "max-age=31536000; includeSubDomains"
	referrerPolicy          = "strict-origin-when-cross-origin"
	permissionsPolicy       = "camera=(), geolocation=(), microphone=()"
)

// SecurityHeaders adds browser security headers to every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("Content-Security-Policy", contentSecurityPolicy)
		header.Set("Strict-Transport-Security", strictTransportSecurity)
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("X-XSS-Protection", "1; mode=block")
		header.Set("Referrer-Policy", referrerPolicy)
		header.Set("Permissions-Policy", permissionsPolicy)
		c.Next()
	}
}
