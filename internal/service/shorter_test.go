package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	config "github.com/lxmp7p/yaGo-url-shortener/internal/config"
	logger "github.com/lxmp7p/yaGo-url-shortener/internal/logger"
	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
)

func TestCreateShortURLBatchAPI_Unauthorized(t *testing.T) {
	s := &ShortenerService{}

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", nil)
	w := httptest.NewRecorder()

	s.CreateShortURLBatchAPI(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestCreateShortURLBatchAPI_InvalidJSON(t *testing.T) {
	s := &ShortenerService{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten/batch",
		strings.NewReader("{invalid}"),
	)

	ctx := context.WithValue(req.Context(), UserIDContextKey(), "user1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	s.CreateShortURLBatchAPI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateShortURLBatchAPI_Empty(t *testing.T) {
	s := &ShortenerService{}

	req := httptest.NewRequest(http.MethodPost, "/api/batch", strings.NewReader("[]"))
	ctx := context.WithValue(req.Context(), UserIDContextKey(), "user1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	s.CreateShortURLBatchAPI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

type storageMock struct {
	saveFn func(ctx context.Context, originalURL, shortURL, userID string) error
}

func (m *storageMock) Save(ctx context.Context, originalURL, shortURL, userID string) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, originalURL, shortURL, userID)
	}
	return nil
}

func (m *storageMock) Get(ctx context.Context, shortURL string) (string, error) {
	return "", nil
}

func (m *storageMock) GetByUserID(ctx context.Context, userID string) ([]repository.URL, error) {
	return nil, nil
}

func (m *storageMock) Delete(ctx context.Context, userID string, IDs []string) error {
	return nil
}

func (m *storageMock) Stats(ctx context.Context) (urls int, users int, err error) {
	return 0, 0, nil
}

func (m *storageMock) Close() error {
	return nil
}

func TestCreateShortURLApi_OK(t *testing.T) {
	s := &ShortenerService{
		Config: config.Config{
			ResultAddr: "http://localhost:8080",
		},
		Storage:    &storageMock{},
		Dispatcher: &logger.Dispatcher{},
	}

	body := `{"url":"https://google.com"}`

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), UserIDContextKey(), "user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	s.CreateShortURLApi(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d", w.Code)
	}
}

func TestCreateShortURLApi_InvalidJSON(t *testing.T) {
	s := &ShortenerService{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader("{"),
	)

	ctx := context.WithValue(req.Context(), UserIDContextKey(), "user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	s.CreateShortURLApi(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
}

func TestCreateShortURLApi_EmptyURL(t *testing.T) {
	s := &ShortenerService{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(`{"url":""}`),
	)

	ctx := context.WithValue(req.Context(), UserIDContextKey(), "user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	s.CreateShortURLApi(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
}

func TestCreateShortURLApi_Unauthorized(t *testing.T) {
	s := &ShortenerService{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		nil,
	)

	w := httptest.NewRecorder()

	s.CreateShortURLApi(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", w.Code)
	}
}

func TestCreateShortURLApi_Retry(t *testing.T) {
	calls := 0

	s := &ShortenerService{
		Config: config.Config{
			ResultAddr: "http://localhost:8080",
		},
		Storage: &storageMock{
			saveFn: func(ctx context.Context, original, short, user string) error {
				calls++
				if calls < 3 {
					return errors.New("collision")
				}
				return nil
			},
		},
		Dispatcher: &logger.Dispatcher{},
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(`{"url":"https://google.com"}`),
	)

	ctx := context.WithValue(req.Context(), UserIDContextKey(), "user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	s.CreateShortURLApi(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d", w.Code)
	}

	if calls != 3 {
		t.Fatalf("expected 3 Save calls got %d", calls)
	}
}

type storageGetMock struct {
	res string
	err error
}

func (m *storageGetMock) Save(ctx context.Context, o, s, u string) error { return nil }
func (m *storageGetMock) Get(ctx context.Context, short string) (string, error) {
	return m.res, m.err
}
func (m *storageGetMock) GetByUserID(ctx context.Context, userID string) ([]repository.URL, error) {
	return nil, nil
}
func (m *storageGetMock) Delete(ctx context.Context, userID string, ids []string) error {
	return nil
}

func (m *storageGetMock) Stats(ctx context.Context) (urls int, users int, err error) {
	return 0, 0, nil
}

func (m *storageGetMock) Close() error { return nil }

type routeCtxKey struct{}

type urlParams struct {
	params map[string]string
}

func (u *urlParams) Get(key string) string {
	return u.params[key]
}

func withURLParam(ctx context.Context, key, value string) context.Context {
	params := make(map[string]string)
	params[key] = value

	return context.WithValue(ctx, chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{key},
			Values: []string{value},
		},
	})
}
func TestGetOriginalURL(t *testing.T) {
	tests := []struct {
		name       string
		userOK     bool
		storageRes string
		storageErr error
		wantCode   int
	}{
		{
			name:     "unauthorized",
			userOK:   false,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:       "not found",
			userOK:     true,
			storageErr: errors.New("not found"),
			wantCode:   http.StatusNotFound,
		},
		{
			name:       "deleted",
			userOK:     true,
			storageErr: repository.ErrURLDeleted,
			wantCode:   http.StatusGone,
		},
		{
			name:       "ok",
			userOK:     true,
			storageRes: "https://google.com",
			wantCode:   http.StatusTemporaryRedirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := &ShortenerService{
				Config: config.Config{
					ResultAddr: "http://localhost",
				},
				Storage: &storageGetMock{
					res: tt.storageRes,
					err: tt.storageErr,
				},
				Dispatcher: &logger.Dispatcher{},
			}

			req := httptest.NewRequest(http.MethodGet, "/x", nil)

			if tt.userOK {
				req = req.WithContext(
					context.WithValue(req.Context(), UserIDContextKey(), "user"),
				)
			}

			// 👉 имитация chi URL param без chi
			req = req.WithContext(
				context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{"short_url"},
						Values: []string{"abc"},
					},
				}),
			)

			w := httptest.NewRecorder()

			s.GetOriginalURL(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("got %d want %d", w.Code, tt.wantCode)
			}
		})
	}
}

type storageCreateMock struct {
	saveFn func() error
}

func (m *storageCreateMock) Save(ctx context.Context, o, s, u string) error {
	if m.saveFn != nil {
		return m.saveFn()
	}
	return nil
}

func (m *storageCreateMock) Get(ctx context.Context, s string) (string, error) {
	return "", nil
}
func (m *storageCreateMock) GetByUserID(ctx context.Context, u string) ([]repository.URL, error) {
	return nil, nil
}
func (m *storageCreateMock) Delete(ctx context.Context, u string, ids []string) error {
	return nil
}

func (m *storageCreateMock) Stats(ctx context.Context) (urls int, users int, err error) {
	return 0, 0, nil
}

func (m *storageCreateMock) Close() error { return nil }

func TestCreateShortURL(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		saveErr     error
		wantCode    int
	}{
		{
			name:        "wrong content type",
			contentType: "application/json",
			body:        "https://google.com",
			wantCode:    http.StatusUnsupportedMediaType,
		},
		{
			name:        "bad body read",
			contentType: "text/plain",
			body:        "",
			saveErr:     nil,
			wantCode:    http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShortenerService{
				Config: config.Config{ResultAddr: "http://localhost"},
				Storage: &storageCreateMock{
					saveFn: func() error {
						return tt.saveErr
					},
				},
				Dispatcher: &logger.Dispatcher{},
			}

			req := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(tt.body))
			req.Header.Set(ContentTypeHeader, tt.contentType)

			req = req.WithContext(
				context.WithValue(req.Context(), UserIDContextKey(), "u1"),
			)

			w := httptest.NewRecorder()

			s.CreateShortURL(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("got %d want %d", w.Code, tt.wantCode)
			}
		})
	}
}
