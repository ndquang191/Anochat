package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCSRFMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		method        string
		origin        string
		allowedOrigin string
		wantStatus    int
	}{
		{
			name:          "allows matching origin",
			method:        http.MethodPost,
			origin:        "http://localhost:3000",
			allowedOrigin: "http://localhost:3000",
			wantStatus:    http.StatusNoContent,
		},
		{
			name:          "allows configured origin with trailing slash",
			method:        http.MethodPut,
			origin:        "https://anochat.example",
			allowedOrigin: "https://anochat.example/",
			wantStatus:    http.StatusNoContent,
		},
		{
			name:          "rejects foreign origin",
			method:        http.MethodDelete,
			origin:        "https://evil.example",
			allowedOrigin: "https://anochat.example",
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "rejects missing origin",
			method:        http.MethodPost,
			allowedOrigin: "https://anochat.example",
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "allows GET without origin",
			method:        http.MethodGet,
			allowedOrigin: "https://anochat.example",
			wantStatus:    http.StatusNoContent,
		},
		{
			name:          "allows HEAD without origin",
			method:        http.MethodHead,
			allowedOrigin: "https://anochat.example",
			wantStatus:    http.StatusNoContent,
		},
		{
			name:          "allows preflight without origin check",
			method:        http.MethodOptions,
			allowedOrigin: "https://anochat.example",
			wantStatus:    http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CSRFMiddleware(tt.allowedOrigin))
			router.Handle(tt.method, "/", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			request := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				request.Header.Set("Origin", tt.origin)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}
