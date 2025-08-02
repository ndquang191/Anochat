package database

import "errors"

var (
	ErrDatabaseURLMissing      = errors.New("DATABASE_URL environment variable is not set")
	ErrDatabaseConfigMissing   = errors.New("required database environment variables (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME) are not set")
	ErrDatabaseNotInitialized  = errors.New("database connection is not initialized")
	ErrUserNotFound           = errors.New("user not found")
	ErrRoomNotFound           = errors.New("room not found")
	ErrMessageNotFound        = errors.New("message not found")
	ErrDuplicateEmail         = errors.New("email already exists")
	ErrInvalidUserID          = errors.New("invalid user ID")
)