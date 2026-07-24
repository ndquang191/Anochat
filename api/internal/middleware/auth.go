package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

func AuthMiddleware(authService *service.AuthService, userRepo repository.UserRepository, cfg *config.Config) gin.HandlerFunc {
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
			tokenString, _ = c.Cookie("access_token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bạn cần đăng nhập để tiếp tục", "code": "no_token"})
			c.Abort()
			return
		}

		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			if strings.Contains(err.Error(), "token is expired") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Phiên đăng nhập đã hết hạn, vui lòng đăng nhập lại", "code": "token_expired"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token xác thực không hợp lệ, vui lòng đăng nhập lại", "code": "token_invalid"})
			}
			c.Abort()
			return
		}

		user, err := userRepo.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy tài khoản", "code": "user_not_found"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("is_admin", user.IsAdmin)
		c.Set("is_active", user.IsActive)
		c.Set("ban_count", user.BanCount)
		c.Set("review_request_count", user.ReviewRequestCount)
		c.Set("review_requested", user.ReviewRequested)
		c.Next()
	}
}

// RequireActive blocks chat and account mutations while still allowing a
// suspended user to load their status and request a review.
func RequireActive() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("is_active") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tài khoản của bạn đã bị khóa", "code": "account_suspended"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAdmin restricts a route to authenticated administrators. It must run
// after AuthMiddleware, which populates the is_admin context value.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("is_admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền thực hiện thao tác này", "code": "admin_required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
