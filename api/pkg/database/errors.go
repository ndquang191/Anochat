package database

import "errors"

var (
	ErrDatabaseURLMissing      = errors.New("database URL is missing")
	ErrDatabaseNotInitialized  = errors.New("database is not initialized")
	ErrUserNotFound           = errors.New("user not found")
	ErrRoomNotFound           = errors.New("room not found")
	ErrMessageNotFound        = errors.New("message not found")
	ErrDuplicateEmail         = errors.New("email already exists")
	ErrInvalidUserID          = errors.New("invalid user ID")
)