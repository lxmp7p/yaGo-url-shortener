package service

import (
	"context"
	"math/rand"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

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
