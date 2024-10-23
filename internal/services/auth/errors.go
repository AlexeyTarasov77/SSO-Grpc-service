package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound = errors.New("app not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrPermissionAlreadyExists = errors.New("permission with this code already exists")
	ErrUserAlreadyActivated = errors.New("user already activated")
	ErrInvalidToken = errors.New("invalid or expired token")
)