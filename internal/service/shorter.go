package service

import (
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
)

const (
	TextContentType   = "text/plain"
	ContentTypeHeader = "Content-Type"
	MaxShortAttempts  = 10
	Chars             = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type ShortenerService struct {
	Config  config.Config
	Storage repository.URLstorage
}

func (s *ShortenerService) Shortenner() (string, error) {
	var randURL string
	var ok bool
	for attempt := 0; attempt < MaxShortAttempts; attempt++ {
		var shortURL string
		for range 8 {
			shortURL += string(Chars[rand.Intn(len(Chars))])
		}
		if !s.Storage.Exist(shortURL) {
			randURL = shortURL
			ok = true
			break
		}
	}
	if !ok {
		return randURL, errors.New("failed to generate short url")
	}
	return randURL, nil
}

func (s *ShortenerService) GetOriginalURL(res http.ResponseWriter, req *http.Request) {
	shortURL := chi.URLParam(req, "short_url")
	originalURL, ok := s.Storage.Get(shortURL)
	if !ok {
		http.Error(res, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *ShortenerService) GetShortURL(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get(ContentTypeHeader)
	if !strings.Contains(strings.ToLower(contentType), TextContentType) {
		http.Error(res, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "failed to parse body", http.StatusBadRequest)
		return
	}

	shortURL, err := s.Shortenner()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Storage.Save(string(body), shortURL)

	res.Header().Set(ContentTypeHeader, TextContentType)
	res.WriteHeader(http.StatusCreated)
	result, err := url.JoinPath(s.Config.ResultAddr, shortURL)
	if err != nil {
		http.Error(res, "failed to parse body", http.StatusBadRequest)
		return
	}
	res.Write([]byte(result))
}
