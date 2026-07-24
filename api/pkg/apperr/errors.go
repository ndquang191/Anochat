package apperr

import "net/http"

// AppError is a user-facing application error that carries an HTTP status code and message.
// Use the sentinel variables below throughout the app; return plain errors for internal failures.
type AppError struct {
	HTTPStatus int
	Message    string
}

func (e *AppError) Error() string { return e.Message }

// Sentinel application errors. Handlers and services return these directly.
var (
	// Auth
	ErrInvalidOAuthState   = &AppError{HTTPStatus: http.StatusBadRequest, Message: "Trạng thái xác thực không hợp lệ, vui lòng thử lại"}
	ErrNoAuthCode          = &AppError{HTTPStatus: http.StatusBadRequest, Message: "Thiếu mã xác thực từ Google"}
	ErrNoRefreshToken      = &AppError{HTTPStatus: http.StatusUnauthorized, Message: "Phiên đăng nhập không tồn tại, vui lòng đăng nhập lại"}
	ErrInvalidRefreshToken = &AppError{HTTPStatus: http.StatusUnauthorized, Message: "Phiên đăng nhập đã hết hạn hoặc không hợp lệ, vui lòng đăng nhập lại"}
	ErrCorruptTokenData    = &AppError{HTTPStatus: http.StatusInternalServerError, Message: "Dữ liệu phiên không hợp lệ, vui lòng đăng nhập lại"}
	ErrInvalidJWT          = &AppError{HTTPStatus: http.StatusUnauthorized, Message: "Token xác thực không hợp lệ, vui lòng đăng nhập lại"}
	ErrUserSuspended       = &AppError{HTTPStatus: http.StatusForbidden, Message: "Tài khoản của bạn đã bị khóa"}

	// User
	ErrUserNotFound    = &AppError{HTTPStatus: http.StatusNotFound, Message: "Không tìm thấy người dùng"}
	ErrProfileNotFound = &AppError{HTTPStatus: http.StatusNotFound, Message: "Hồ sơ không tồn tại hoặc đã bị ẩn"}
	ErrUnauthenticated = &AppError{HTTPStatus: http.StatusUnauthorized, Message: "Bạn cần đăng nhập để thực hiện thao tác này"}
	ErrInvalidDisplayName = &AppError{HTTPStatus: http.StatusBadRequest, Message: "Tên hiển thị phải có từ 2 đến 32 ký tự"}
	ErrDisplayNameCooldown = &AppError{HTTPStatus: http.StatusTooManyRequests, Message: "Tên hiển thị chỉ có thể thay đổi một lần trong 30 ngày"}

	// Room
	ErrNoActiveRoom  = &AppError{HTTPStatus: http.StatusNotFound, Message: "Bạn hiện không có phòng chat nào đang hoạt động"}
	ErrNotRoomMember = &AppError{HTTPStatus: http.StatusForbidden, Message: "Bạn không phải thành viên của phòng chat này"}

	// Queue
	ErrHasActiveRoom  = &AppError{HTTPStatus: http.StatusBadRequest, Message: "Bạn đang có phòng chat đang hoạt động, vui lòng rời phòng trước khi tham gia hàng chờ"}
	ErrAlreadyInQueue = &AppError{HTTPStatus: http.StatusBadRequest, Message: "Bạn đã ở trong hàng chờ rồi"}
	ErrNotInQueue     = &AppError{HTTPStatus: http.StatusBadRequest, Message: "Bạn không có trong hàng chờ"}

	// Input / general
	ErrInvalidBody = &AppError{HTTPStatus: http.StatusBadRequest, Message: "Dữ liệu gửi lên không hợp lệ"}
	ErrInvalidID   = &AppError{HTTPStatus: http.StatusBadRequest, Message: "ID không hợp lệ"}
	ErrForbidden   = &AppError{HTTPStatus: http.StatusForbidden, Message: "Bạn không có quyền thực hiện thao tác này"}
)
