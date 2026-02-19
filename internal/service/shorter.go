package service

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

const (
	TextContentType   = "text/plain"
	ContentTypeHeader = "Content-Type"
)

var urlCache = make(map[string]string)
var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Shortenner() string {
	var randUrl string
	for {
		var shortUrl string
		for range 8 {
			shortUrl += string(chars[rand.Intn(len(chars))])
		}
		_, exist := urlCache[shortUrl]
		if !exist {
			randUrl = shortUrl
			break
		}
	}
	return randUrl
}

func GetOriginalUrl(res http.ResponseWriter, req *http.Request) {
	shortUrl := req.URL.Path[1:]
	originalUrl, exist := urlCache[shortUrl]
	if !exist {
		http.Error(res, "URL not found", http.StatusNotFound)
		return
	}
	res.Header().Set("Location", originalUrl)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func GetShortUrl(res http.ResponseWriter, req *http.Request) {
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

	originalUrl := string(body)
	shortUrl := Shortenner()
	urlCache[shortUrl] = originalUrl

	res.Header().Set(ContentTypeHeader, TextContentType)
	res.WriteHeader(http.StatusCreated)

	result := "http" + "://" + req.Host + "/" + shortUrl
	res.Write([]byte(result))
}
