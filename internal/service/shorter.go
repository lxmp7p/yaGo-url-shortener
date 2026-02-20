package service

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

const (
	TextContentType   = "text/plain"
	ContentTypeHeader = "Content-Type"
)

var urlCache = make(map[string]string)
var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Shortenner() string {
	var randURL string
	for {
		var shortURL string
		for range 8 {
			shortURL += string(chars[rand.Intn(len(chars))])
		}
		_, exist := urlCache[shortURL]
		if !exist {
			randURL = shortURL
			break
		}
	}
	return randURL
}

func GetOriginalURL(res http.ResponseWriter, req *http.Request) {
	shortURL := chi.URLParam(req, "short_url")
	originalURL, exist := urlCache[shortURL]
	if !exist {
		http.Error(res, "URL not found", http.StatusNotFound)
		return
	}
	res.Header().Set("Location", originalURL)
	fmt.Println(originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func GetShortURL(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get(ContentTypeHeader)
	if !strings.Contains(strings.ToLower(contentType), TextContentType) {
		http.Error(res, "method not allowed", http.StatusUnsupportedMediaType)
		return
	}

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "failed to parse body", http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	shortURL := Shortenner()
	urlCache[shortURL] = originalURL

	res.Header().Set(ContentTypeHeader, TextContentType)
	res.WriteHeader(http.StatusCreated)

	result := "http" + "://" + req.Host + "/" + shortURL
	res.Write([]byte(result))
}
