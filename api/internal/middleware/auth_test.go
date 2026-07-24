package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireActive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		isActive   bool
		wantStatus int
	}{
		{name: "active user", isActive: true, wantStatus: http.StatusNoContent},
		{name: "suspended user", isActive: false, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("is_active", tt.isActive)
			})
			router.Use(RequireActive())
			router.GET("/", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		isAdmin    bool
		wantStatus int
	}{
		{name: "admin user", isAdmin: true, wantStatus: http.StatusNoContent},
		{name: "non-admin user", isAdmin: false, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("is_admin", tt.isAdmin)
			})
			router.Use(RequireAdmin())
			router.GET("/", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}
