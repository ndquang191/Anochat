package middleware

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
)

// CSRFMiddleware protects cookie-authenticated state-changing requests by
// requiring the browser-supplied Origin header to match the configured client.
func CSRFMiddleware(allowedOrigin string) gin.HandlerFunc {
	allowedOrigin = strings.TrimSuffix(allowedOrigin, "/")

	return func(c *gin.Context) {
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		origin := strings.TrimSuffix(c.GetHeader("Origin"), "/")
		if origin == "" || origin != allowedOrigin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Nguồn yêu cầu không hợp lệ",
				"code":  "invalid_origin",
			})
			return
		}

		c.Next()
	}
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
