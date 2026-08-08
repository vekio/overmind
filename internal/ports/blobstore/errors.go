package blobstore

import "errors"

var (
	ErrAlreadyExists = errors.New("blob already exists")
	ErrNotFound      = errors.New("blob not found")
	ErrInvalidPath   = errors.New("invalid blob path")
)
