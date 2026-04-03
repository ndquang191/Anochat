package dto

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
)

// ApiResponse is a generic API response envelope.
type ApiResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// OK sends a 200 success response with data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    data,
	})
}

// OKWithMessage sends a 200 success response with a message and optional data.
func OKWithMessage(c *gin.Context, message string, data any) {
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

// FailErr sends an error response derived from err.
// If err is an *apperr.AppError, its HTTP status and message are used.
// Otherwise a 500 is returned and the error is logged server-side.
func FailErr(c *gin.Context, err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		Fail(c, appErr.HTTPStatus, appErr.Message)
		return
	}
	slog.Error("unhandled internal error", "error", err)
	Fail(c, http.StatusInternalServerError, "Đã xảy ra lỗi hệ thống, vui lòng thử lại sau")
}
