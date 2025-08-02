package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/service"
)

// UserHandler handles user-related endpoints
type UserHandler struct {
	authService *service.AuthService
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(authService *service.AuthService, userService *service.UserService) *UserHandler {
	return &UserHandler{
		authService: authService,
		userService: userService,
	}
}

// GetUserState gets user state including active room and messages
func (h *UserHandler) GetUserState(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get JWT token from header for AuthService
	authHeader := c.GetHeader("Authorization")
	tokenString := authHeader[7:] // Remove "Bearer " prefix

	// Get user state using AuthService
	user, room, messages, err := h.authService.GetUserFromToken(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if user is new (no profile)
	profile, err := h.userService.GetProfile(c.Request.Context(), userID)
	isNewUser := err != nil || profile == nil

	// Build response
	response := gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"avatar_url": user.AvatarURL,
		},
		"room":        nil,
		"messages":    nil,
		"is_new_user": isNewUser,
	}

	// Add profile if exists
	if profile != nil {
		response["user"].(gin.H)["profile"] = gin.H{
			"age":       profile.Age,
			"city":      profile.City,
			"is_male":   profile.IsMale,
			"is_hidden": profile.IsHidden,
		}
	}

	// Add room and messages if exists
	if room != nil {
		response["room"] = gin.H{
			"id":       room.ID,
			"user1_id": room.User1ID,
			"user2_id": room.User2ID,
			"category": room.Category,
		}
		response["messages"] = messages
	}

	c.JSON(http.StatusOK, response)
}
