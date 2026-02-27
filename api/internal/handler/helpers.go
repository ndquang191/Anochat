package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

func getUserID(c *gin.Context) uuid.UUID {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}
	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return userID
}

func signOutAndRedirect(c *gin.Context, cfg *config.Config) {
	c.SetCookie("jwt_token", "", -1, "/", "", cfg.IsProduction(), true)
	c.Redirect(http.StatusTemporaryRedirect, cfg.ClientURL+"/login")
}
