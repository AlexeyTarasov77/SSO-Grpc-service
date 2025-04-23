package permissions

import "errors"

var (
	ErrPermissionAlreadyExists = errors.New("permission with this code already exists")
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrUserNotFound            = errors.New("Related user not found")
)
