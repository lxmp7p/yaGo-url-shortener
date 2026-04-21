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
	"time"

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

type DeleteTask struct {
	UserID string
	IDs    []string
}

type URLstorage interface {
	Save(ctx context.Context, originalURL string, shortURL string, userID string) error
	Get(ctx context.Context, shortURL string) (string, error)
	GetByUserID(ctx context.Context, userID string) ([]repository.URL, error)
	Delete(ctx context.Context, userID string, IDs []string) error
}

type ShortenerService struct {
	Config   config.Config
	Storage  URLstorage
	Logger   logrus.Logger
	Database *sql.DB
	secret   []byte
	DeleteCh chan DeleteTask
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

func (s *ShortenerService) CreateShortURLBatchAPI(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var request []Original

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(request) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	var response []Shorten

	for _, item := range request {
		var shortURL string
		var err error

		for attempt := 0; attempt < MaxShortAttempts; attempt++ {
			shortURL = generateShortURL()
			err := s.Storage.Save(r.Context(), item.OriginalURL, shortURL, userID)
			if err != nil {
				continue
			}
			break
		}

		shortURL, err = url.JoinPath(s.Config.ResultAddr, shortURL)
		if err != nil {
			http.Error(w, "failed to generate URL", http.StatusBadRequest)
			return
		}

		response = append(response, Shorten{
			ID:       item.ID,
			ShortURL: shortURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (s *ShortenerService) CreateShortURLApi(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var shortenRequest ShortenRequest
	status := http.StatusCreated

	if err := json.NewDecoder(r.Body).Decode(&shortenRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := shortenRequest.Bind(r); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var shortURL string
	for attempt := 0; attempt < MaxShortAttempts; attempt++ {
		shortURL = generateShortURL()
		err := s.Storage.Save(r.Context(), string(shortenRequest.URL), shortURL, userID)
		if err != nil {
			var URLErr *repository.URLError
			if errors.As(err, &URLErr) {
				w.Header().Set("Content-Type", "application/json")
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
		http.Error(w, "failed to generate URL", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ShortenResponse{Result: shortURL})
}

func (s *ShortenerService) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "short_url")
	originalURL, err := s.Storage.Get(r.Context(), shortURL)
	if err != nil {
		slog.Error(err.Error())
		if errors.Is(err, repository.ErrURLDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *ShortenerService) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	userID, _ := UserIDFromContext(r.Context())

	contentType := r.Header.Get(ContentTypeHeader)
	status := http.StatusCreated

	if !strings.Contains(strings.ToLower(contentType), TextContentType) {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}

	var shortURL string
	for attempt := 0; attempt < MaxShortAttempts; attempt++ {
		shortURL = generateShortURL()
		err = s.Storage.Save(r.Context(), string(body), shortURL, userID)
		if err != nil {
			var URLErr *repository.URLError
			if errors.As(err, &URLErr) {
				w.Header().Set("Content-Type", "application/json")
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(ContentTypeHeader, TextContentType)
	w.WriteHeader(status)
	result, err := url.JoinPath(s.Config.ResultAddr, shortURL)
	if err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}
	w.Write([]byte(result))
}

func (s *ShortenerService) GetUsersURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	URLs, err := s.Storage.GetByUserID(r.Context(), userID)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for i, u := range URLs {
		shortURL, err := url.JoinPath(s.Config.ResultAddr, u.Short)
		if err == nil {
			URLs[i].Short = shortURL
		}
	}

	if len(URLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set(ContentTypeHeader, "application/json")
	json.NewEncoder(w).Encode(URLs)
}

func (s *ShortenerService) DeleteUsersURLs(w http.ResponseWriter, r *http.Request) {
	userID, _ := UserIDFromContext(r.Context())

	var ids []string
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	s.DeleteCh <- DeleteTask{
		UserID: userID,
		IDs:    ids,
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *ShortenerService) StartDeleteWorker() {
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		batch := make(map[string][]string)

		for {
			select {
			case task := <-s.DeleteCh:
				batch[task.UserID] = append(batch[task.UserID], task.IDs...)
			case <-ticker.C:
				for userID, ids := range batch {
					if len(ids) > 0 {
						err := s.Storage.Delete(context.Background(), userID, ids)
						s.Logger.Error(err)
					}
				}
				batch = make(map[string][]string)
			}
		}
	}()
}
