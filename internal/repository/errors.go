package repository

import "errors"

var (
	urlNotFound      = "url not found"
	OriginalUrlExist = errors.New("original url exist")
)
