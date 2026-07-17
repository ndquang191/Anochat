package initialize

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	healthy := func() error { return nil }
	unhealthy := func() error { return errors.New("unavailable") }

	tests := []struct {
		name          string
		databaseCheck func() error
		redisCheck    func() error
		wantStatus    int
	}{
		{name: "healthy", databaseCheck: healthy, redisCheck: healthy, wantStatus: http.StatusOK},
		{name: "database unavailable", databaseCheck: unhealthy, redisCheck: healthy, wantStatus: http.StatusServiceUnavailable},
		{name: "redis unavailable", databaseCheck: healthy, redisCheck: unhealthy, wantStatus: http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/healthz", healthHandlerWithChecks(tt.databaseCheck, tt.redisCheck))
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}
