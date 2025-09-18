package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/internal/util"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"golang.org/x/oauth2"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
	oauthConfig *oauth2.Config
	config      *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, oauthConfig *oauth2.Config, config *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		oauthConfig: oauthConfig,
		config:      config,
	}
}

// GoogleLogin redirects to Google OAuth
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := h.oauthConfig.AuthCodeURL("random-state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles OAuth callback and returns JWT token
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Get authorization code from query params
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code required"})
		return
	}

	// Process OAuth callback
	user, jwtToken, err := h.authService.ProcessOAuthCallback(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set JWT token in HTTP-only cookie
	c.SetCookie("jwt_token", jwtToken, 3600*24*7, "/", "", false, true) // 7 days, HTTP-only, no domain restriction

	// Create user data for frontend
	userData := gin.H{
		"id":         user.ID,
		"email":      *user.Email,
		"name":       *user.Name,
		"avatar_url": *user.AvatarURL,
	}

	// Set user data in a temporary cookie (will be cleared by frontend)
	userDataJSON, _ := json.Marshal(userData)
	c.SetCookie("temp_user_data", string(userDataJSON), 60, "/", "", false, false) // 1 minute, not HTTP-only, no domain restriction

	// Redirect to frontend callback page
	frontendURL := h.config.ClientURL + "/callback"
	c.Redirect(http.StatusTemporaryRedirect, frontendURL)
}

// Logout clears the JWT token cookie
func (h *AuthHandler) Logout(c *gin.Context) {
	// Optional redirect target via query param, fallback to CLIENT_URL + /login
	redirectURL := c.Query("redirect")
	if redirectURL != "" {
		util.SignOutAndRedirect(c, redirectURL)
	} else {
		util.SignOutAndRedirect(c)
	}
}
