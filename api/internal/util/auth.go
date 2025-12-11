package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

// SignOutAndRedirect clears the JWT cookie and redirects to the given URL.
// If redirectURL is empty, defaults to CLIENT_URL + "/login" from environment.
// Reuse this in middleware/handlers when authorization fails or user logs out.
func SignOutAndRedirect(c *gin.Context, redirectURL ...string) {
	// Clear JWT token cookie by setting it to expire immediately
	c.SetCookie("jwt_token", "", -1, "/", "", false, true)

	// Determine redirect URL
	var finalRedirectURL string
	if len(redirectURL) > 0 && redirectURL[0] != "" {
		finalRedirectURL = redirectURL[0]
	} else {
		// Load config to get CLIENT_URL
		cfg := config.Load()
		finalRedirectURL = cfg.ClientURL + "/login"
	}

	// Perform HTTP redirect
	c.Redirect(http.StatusTemporaryRedirect, finalRedirectURL)
}
