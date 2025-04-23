package storage

import "errors"

var (
	ErrRecordNotFound      = errors.New("record not found")
	ErrRecordAlreadyExists = errors.New("record already exists")
)

