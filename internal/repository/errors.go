package repository

import "errors"

var (
	ErrURLNotFound       = errors.New("url not found")
	ErrURLDeleted        = errors.New("url deleted")
	ErrOriginalURLExists = errors.New("original url exist")
)
