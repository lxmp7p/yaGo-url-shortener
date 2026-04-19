package repository

import "errors"

var (
	ErrURLNotFound       = errors.New("url not found")
	ErrOriginalURLExists = errors.New("original url exist")
)
