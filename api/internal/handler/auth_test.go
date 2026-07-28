package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestGoogleCallbackRedirectsOAuthValidationErrorsToClient(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		target       string
		oauthState   string
		wantLocation string
	}{
		{
			name:         "invalid state",
			target:       "/auth/callback?state=unexpected",
			oauthState:   "expected",
			wantLocation: "https://chat.example.com/error?error=invalid_state",
		},
		{
			name:         "missing code",
			target:       "/auth/callback?state=expected",
			oauthState:   "expected",
			wantLocation: "https://chat.example.com/error?error=missing_code",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authHandler := &AuthHandler{
				config: &config.Config{ClientURL: "https://chat.example.com/"},
			}
			router := gin.New()
			router.GET("/auth/callback", authHandler.GoogleCallback)

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.target, nil)
			request.AddCookie(&http.Cookie{Name: "oauth_state", Value: test.oauthState})
			router.ServeHTTP(response, request)

			assert.Equal(t, http.StatusSeeOther, response.Code)
			assert.Equal(t, test.wantLocation, response.Header().Get("Location"))
		})
	}
}
