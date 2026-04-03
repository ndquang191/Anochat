package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
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
		dto.FailErr(c, apperr.ErrInvalidOAuthState)
		return
	}
	c.SetCookie("oauth_state", "", -1, "/", "", h.config.IsProduction(), true)

	code := c.Query("code")
	if code == "" {
		dto.FailErr(c, apperr.ErrNoAuthCode)
		return
	}

	result, err := h.authService.ProcessOAuthCallback(c.Request.Context(), code)
	if err != nil {
		dto.FailErr(c, err)
		return
	}

	h.setAccessTokenCookie(c, result.AccessToken)
	h.setRefreshTokenCookie(c, result.RefreshToken)

	userData := gin.H{
		"id":         result.User.ID,
		"email":      *result.User.Email,
		"name":       *result.User.Name,
		"avatar_url": *result.User.AvatarURL,
	}
	userDataJSON, _ := json.Marshal(userData)
	c.SetCookie("temp_user_data", string(userDataJSON), 60, "/", "", h.config.IsProduction(), false)

	frontendURL := h.config.ClientURL + "/callback"
	c.Redirect(http.StatusTemporaryRedirect, frontendURL)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		dto.FailErr(c, apperr.ErrNoRefreshToken)
		return
	}

	accessToken, err := h.authService.RefreshAccessToken(c.Request.Context(), refreshToken)
	if err != nil {
		// Clear invalid refresh token cookie
		h.clearRefreshTokenCookie(c)
		dto.FailErr(c, err)
		return
	}

	h.setAccessTokenCookie(c, accessToken)
	dto.OK(c, gin.H{"message": "Token refreshed"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Revoke refresh token from Redis
	if refreshToken, err := c.Cookie("refresh_token"); err == nil {
		h.authService.RevokeRefreshToken(c.Request.Context(), refreshToken)
	}

	h.clearAccessTokenCookie(c)
	h.clearRefreshTokenCookie(c)

	redirectURL := c.Query("redirect")
	if redirectURL == "" {
		redirectURL = h.config.ClientURL + "/login"
	}
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// --- Cookie helpers ---

func (h *AuthHandler) setAccessTokenCookie(c *gin.Context, token string) {
	maxAge := int(h.config.OAuth.AccessTokenExpiry.Seconds())
	c.SetCookie("access_token", token, maxAge, "/", "", h.config.IsProduction(), true)
}

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	maxAge := int(h.config.OAuth.RefreshTokenExpiry.Seconds())
	c.SetCookie("refresh_token", token, maxAge, "/auth", "", h.config.IsProduction(), true)
}

func (h *AuthHandler) clearAccessTokenCookie(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", h.config.IsProduction(), true)
	c.SetCookie("jwt_token", "", -1, "/", "", h.config.IsProduction(), true) // clear legacy cookie
}

func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/auth", "", h.config.IsProduction(), true)
}

func generateOAuthState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
