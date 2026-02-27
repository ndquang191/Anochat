package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

func AuthMiddleware(authService *service.AuthService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				tokenString = tokenParts[1]
			}
		}

		if tokenString == "" {
			tokenString, _ = c.Cookie("jwt_token")
		}

		if tokenString == "" {
			c.SetCookie("jwt_token", "", -1, "/", "", cfg.IsProduction(), true)
			c.Redirect(http.StatusTemporaryRedirect, cfg.ClientURL+"/login")
			c.Abort()
			return
		}

		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			c.SetCookie("jwt_token", "", -1, "/", "", cfg.IsProduction(), true)
			c.Redirect(http.StatusTemporaryRedirect, cfg.ClientURL+"/login")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}
