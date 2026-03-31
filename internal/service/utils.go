package service

import "math/rand"

func generateShortURL() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = Chars[rand.Intn(len(Chars))]
	}

	return string(b)
}

