package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SignOutAndRedirect clears the JWT cookie and redirects to the given URL.
// Reuse this in middleware/handlers when authorization fails or user logs out.
func SignOutAndRedirect(c *gin.Context, redirectURL string) {
	// Clear JWT token cookie by setting it to expire immediately
	c.SetCookie("jwt_token", "", -1, "/", "", false, true)

	// Perform HTTP redirect
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

