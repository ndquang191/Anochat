package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGoogleLoginPromptsAccountSelection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authHandler := &AuthHandler{
		oauthConfig: &oauth2.Config{
			ClientID:    "client-id",
			RedirectURL: "https://api.example.com/auth/callback",
			Scopes:      []string{"openid", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL: "https://accounts.google.com/o/oauth2/v2/auth",
			},
		},
		config: &config.Config{},
	}
	router := gin.New()
	router.GET("/auth/google", authHandler.GoogleLogin)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/google", nil)
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusTemporaryRedirect, response.Code)
	location, err := url.Parse(response.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "select_account", location.Query().Get("prompt"))
	assert.NotEmpty(t, location.Query().Get("state"))
}

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
