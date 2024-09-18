package storage

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrAppNotFound = errors.New("app not found")
	ErrAppAlreadyExists = errors.New("app already exists")
	ErrTokenNotFound = errors.New("token not found")
)