package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
	"github.com/sirupsen/logrus"
)

const (
	TextContentType   = "text/plain"
	ContentTypeHeader = "Content-Type"
	MaxShortAttempts  = 10
	Chars             = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type URLstorage interface {
	Save(ctx context.Context, originalURL string, shortURL string) error
	Get(ctx context.Context, shortURL string) (string, error)
}

type ShortenerService struct {
	Config   config.Config
	Storage  URLstorage
	Logger   logrus.Logger
	Database *sql.DB
	secret   []byte
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

type Original struct {
	ID          string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
}

type Shorten struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

func (s *ShortenerService) CreateShortURLBatchAPI(res http.ResponseWriter, req *http.Request) {
	var request []Original

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(res, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(request) == 0 {
		http.Error(res, "Empty batch", http.StatusBadRequest)
	}

	var response []Shorten

	for _, item := range request {
		var shortURL string
		var err error

		for attempt := 0; attempt < MaxShortAttempts; attempt++ {
			shortURL = generateShortURL()
			err := s.Storage.Save(req.Context(), item.OriginalURL, shortURL)
			if err != nil {
				continue
			}
			break
		}

		shortURL, err = url.JoinPath(s.Config.ResultAddr, shortURL)
		if err != nil {
			http.Error(res, "failed to generate URL", http.StatusBadRequest)
			return
		}

		response = append(response, Shorten{
			ID:       item.ID,
			ShortURL: shortURL,
		})
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	json.NewEncoder(res).Encode(response)
}

func (s *ShortenerService) CreateShortURLApi(res http.ResponseWriter, req *http.Request) {
	var shortenRequest ShortenRequest
	status := http.StatusCreated

	if err := json.NewDecoder(req.Body).Decode(&shortenRequest); err != nil {
		http.Error(res, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := shortenRequest.Bind(req); err != nil {
		http.Error(res, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var shortURL string
	for attempt := 0; attempt < MaxShortAttempts; attempt++ {
		shortURL = generateShortURL()
		err := s.Storage.Save(req.Context(), string(shortenRequest.URL), shortURL)
		if err != nil {
			var URLErr *repository.URLError
			if errors.As(err, &URLErr) {
				res.Header().Set("Content-Type", "application/json")
				status = http.StatusConflict
				shortURL = URLErr.Short
				err = nil
				break
			}
			continue
		}
		break
	}

	shortURL, err := url.JoinPath(s.Config.ResultAddr, shortURL)
	if err != nil {
		http.Error(res, "failed to generate URL", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	json.NewEncoder(res).Encode(ShortenResponse{Result: shortURL})
}

func (s *ShortenerService) GetOriginalURL(res http.ResponseWriter, req *http.Request) {
	shortURL := chi.URLParam(req, "short_url")
	originalURL, err := s.Storage.Get(req.Context(), shortURL)
	if err != nil {
		slog.Error(err.Error())
		http.Error(res, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *ShortenerService) CreateShortURL(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get(ContentTypeHeader)
	status := http.StatusCreated

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
		err = s.Storage.Save(req.Context(), string(body), shortURL)
		if err != nil {
			var URLErr *repository.URLError
			if errors.As(err, &URLErr) {
				res.Header().Set("Content-Type", "application/json")
				status = http.StatusConflict
				shortURL = URLErr.Short
				err = nil
				break
			}
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
	res.WriteHeader(status)
	result, err := url.JoinPath(s.Config.ResultAddr, shortURL)
	if err != nil {
		http.Error(res, "failed to parse body", http.StatusBadRequest)
		return
	}
	res.Write([]byte(result))
}

// func (s *ShortenerService) GetUsersURLs(w http.ResponseWriter, r *http.Request) {
// 	userID, ok := r.Context().Value("userID").(string)
// 	if !ok {
// 		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
// 		return
// 	}

// 	contentType := req.Header.Get(ContentTypeHeader)
// 	status := http.StatusCreated

// 	if !strings.Contains(strings.ToLower(contentType), TextContentType) {
// 		http.Error(res, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
// 		return
// 	}

// 	defer req.Body.Close()
// 	body, err := io.ReadAll(req.Body)
// 	if err != nil {
// 		http.Error(res, "failed to parse body", http.StatusBadRequest)
// 		return
// 	}

// 	var shortURL string
// 	for attempt := 0; attempt < MaxShortAttempts; attempt++ {
// 		shortURL = generateShortURL()
// 		err = s.Storage.Save(req.Context(), string(body), shortURL)
// 		if err != nil {
// 			var URLErr *repository.URLError
// 			if errors.As(err, &URLErr) {
// 				res.Header().Set("Content-Type", "application/json")
// 				status = http.StatusConflict
// 				shortURL = URLErr.Short
// 				err = nil
// 				break
// 			}
// 			continue
// 		}
// 		break
// 	}

// 	if err != nil {
// 		slog.Error(err.Error())
// 		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
// 		return
// 	}

// 	res.Header().Set(ContentTypeHeader, TextContentType)
// 	res.WriteHeader(status)
// 	result, err := url.JoinPath(s.Config.ResultAddr, shortURL)
// 	if err != nil {
// 		http.Error(res, "failed to parse body", http.StatusBadRequest)
// 		return
// 	}
// 	res.Write([]byte(result))
// }
