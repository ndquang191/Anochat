package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	authService *service.AuthService
	oauthConfig *oauth2.Config
	config      *config.Config
}

func NewAuthHandler(authService *service.AuthService, oauthConfig *oauth2.Config, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		oauthConfig: oauthConfig,
		config:      cfg,
	}
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := generateOAuthState()
	c.SetCookie("oauth_state", state, 300, "/", "", h.config.IsProduction(), true)
	url := h.oauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	savedState, err := c.Cookie("oauth_state")
	if err != nil || state != savedState {
		dto.Fail(c, http.StatusBadRequest, "Invalid OAuth state")
		return
	}
	c.SetCookie("oauth_state", "", -1, "/", "", h.config.IsProduction(), true)

	code := c.Query("code")
	if code == "" {
		dto.Fail(c, http.StatusBadRequest, "Authorization code required")
		return
	}

	user, jwtToken, err := h.authService.ProcessOAuthCallback(c.Request.Context(), code)
	if err != nil {
		dto.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.SetCookie("jwt_token", jwtToken, 3600*24*7, "/", "", h.config.IsProduction(), true)

	userData := gin.H{
		"id":         user.ID,
		"email":      *user.Email,
		"name":       *user.Name,
		"avatar_url": *user.AvatarURL,
	}
	userDataJSON, _ := json.Marshal(userData)
	c.SetCookie("temp_user_data", string(userDataJSON), 60, "/", "", h.config.IsProduction(), false)

	frontendURL := h.config.ClientURL + "/callback"
	c.Redirect(http.StatusTemporaryRedirect, frontendURL)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("jwt_token", "", -1, "/", "", h.config.IsProduction(), true)

	redirectURL := c.Query("redirect")
	if redirectURL == "" {
		redirectURL = h.config.ClientURL + "/login"
	}
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func generateOAuthState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
