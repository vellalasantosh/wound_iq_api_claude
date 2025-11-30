package utils

import "errors"

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrUnauthorized       = errors.New("unauthorized access")

	// Password errors
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
	ErrPasswordTooLong  = errors.New("password must not exceed 100 characters")
	ErrPasswordMismatch = errors.New("passwords do not match")

	// General errors
	ErrInternalServer = errors.New("internal server error")
	ErrBadRequest     = errors.New("bad request")
	ErrNotFound       = errors.New("resource not found")
)
