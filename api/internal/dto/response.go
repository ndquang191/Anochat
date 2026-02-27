package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ApiResponse is a generic API response envelope.
type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// OK sends a 200 success response with data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    data,
	})
}

// OKWithMessage sends a 200 success response with a message and optional data.
func OKWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Fail sends an error response with the given status code and message.
func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, ApiResponse{
		Success: false,
		Error:   message,
	})
}

// FailWithMessage sends an error response with separate error and human message.
func FailWithMessage(c *gin.Context, status int, errMsg, message string) {
	c.JSON(status, ApiResponse{
		Success: false,
		Error:   errMsg,
		Message: message,
	})
}
