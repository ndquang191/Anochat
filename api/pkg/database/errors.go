package database

import "errors"

// Database connection errors
var (
	ErrDatabaseConfigMissing  = errors.New("required database environment variables (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME) are not set")
	ErrDatabaseNotInitialized = errors.New("database connection is not initialized")
)
