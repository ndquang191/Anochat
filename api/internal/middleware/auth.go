package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/internal/util"
)

// AuthMiddleware validates JWT token and injects user into context
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// First try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Check Bearer token format
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				tokenString = tokenParts[1]
			}
		}

		// If no token from header, try to get from HTTP-only cookie
		if tokenString == "" {
			tokenString, _ = c.Cookie("jwt_token")
		}

		// If still no token, sign out and redirect to login page
		if tokenString == "" {
			println("No token found")
			util.SignOutAndRedirect(c)
			c.Abort()
			return
		}

		// Validate JWT token
		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			util.SignOutAndRedirect(c)
			c.Abort()
			return
		}

		// Inject user ID into context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
