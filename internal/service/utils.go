package service

import (
	"math/rand"
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "userID"

func generateShortURL() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = Chars[rand.Intn(len(Chars))]
	}

	return string(b)
}

func GenerateUserCookie(userID, signature string) *http.Cookie {
	return &http.Cookie{
		Name:     "user",
		Value:    userID + ":" + signature,
		Path:     "/",
		HttpOnly: true,
	}
}
