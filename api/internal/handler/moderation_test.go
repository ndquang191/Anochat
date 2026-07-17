package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireAdminUsesAuthenticatedRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ModerationHandler{}

	tests := []struct {
		name        string
		isAdmin     bool
		wantAllowed bool
		wantStatus  int
	}{
		{name: "admin", isAdmin: true, wantAllowed: true, wantStatus: http.StatusOK},
		{name: "regular user", isAdmin: false, wantAllowed: false, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Set("is_admin", tt.isAdmin)

			if allowed := handler.requireAdmin(context); allowed != tt.wantAllowed {
				t.Errorf("requireAdmin() = %v, want %v", allowed, tt.wantAllowed)
			}
			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}
