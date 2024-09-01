package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)