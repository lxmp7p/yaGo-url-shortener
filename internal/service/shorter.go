package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/sirupsen/logrus"
)

const (
	TextContentType   = "text/plain"
	ContentTypeHeader = "Content-Type"
	MaxShortAttempts  = 10
	Chars             = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type URLstorage interface {
	Save(originalURL string, shortURL string) error
	Get(shortURL string) (string, error)
}

type ShortenerService struct {
	Config  config.Config
	Storage URLstorage
	Logger  logrus.Logger
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func (sr *ShortenRequest) Bind(r *http.Request) error {
	if sr.URL == "" {
		return fmt.Errorf("URL empty")
	}
	return nil
}

func (s *ShortenerService) GetShortURLApi(res http.ResponseWriter, req *http.Request) {
	var shortenRequest ShortenRequest
	fmt.Println("1")

	if err := json.NewDecoder(req.Body).Decode(&shortenRequest); err != nil {
		fmt.Println("2")
		http.Error(res, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := shortenRequest.Bind(req); err != nil {
		fmt.Println("3")

		http.Error(res, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Println("4")

	var shortURL string
	for attempt := 0; attempt < MaxShortAttempts; attempt++ {
		shortURL = generateShortURL()
		err := s.Storage.Save(string(shortenRequest.URL), shortURL)
		if err != nil {
			continue
		}
		break
	}
	fmt.Println("5")

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	json.NewEncoder(res).Encode(ShortenResponse{Result: shortURL})
}

func (s *ShortenerService) GetOriginalURL(res http.ResponseWriter, req *http.Request) {
	shortURL := chi.URLParam(req, "short_url")
	originalURL, err := s.Storage.Get(shortURL)
	if err != nil {
		slog.Error(err.Error())
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

	var shortURL string
	for attempt := 0; attempt < MaxShortAttempts; attempt++ {
		shortURL = generateShortURL()
		err = s.Storage.Save(string(body), shortURL)
		if err != nil {
			continue
		}
		break
	}

	if err != nil {
		slog.Error(err.Error())
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set(ContentTypeHeader, TextContentType)
	res.WriteHeader(http.StatusCreated)
	result, err := url.JoinPath(s.Config.ResultAddr, shortURL)
	if err != nil {
		http.Error(res, "failed to parse body", http.StatusBadRequest)
		return
	}
	res.Write([]byte(result))
}
