package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/internal/service"
)

// AuthMiddleware validates JWT token and injects user into context
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Debug: Log all cookies
		cookies := c.Request.Cookies()
		for _, cookie := range cookies {
			println("Cookie:", cookie.Name, "=", cookie.Value)
		}

		// First try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Check Bearer token format
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				tokenString = tokenParts[1]
				println("Token from Authorization header:", tokenString)
			}
		}

		// If no token from header, try to get from HTTP-only cookie
		if tokenString == "" {
			tokenString, _ = c.Cookie("jwt_token")
			println("Token from cookie:", tokenString)
		}

		// If still no token, return unauthorized
		if tokenString == "" {
			println("No token found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		// Validate JWT token
		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Inject user ID into context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
