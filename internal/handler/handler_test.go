package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	logs "github.com/lxmp7p/yaGo-url-shortener/internal/logger"
	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShortURLHandler(t *testing.T) {
	var tmpURL string
	var templateURL = "http://localhost:8080/"

	tmpFile, err := os.CreateTemp("", "tmp-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	app := App{
		Config: config.Config{
			Addr:       "localhost:8080",
			ResultAddr: "http://localhost:8080",
		},
		Storage: repository.NewCache(context.Background(), tmpFile.Name(), tmpFile),
		Logger:  logrus.New(),
	}

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name        string
		method      string
		path        string
		body        *bytes.Reader
		contentType string
		want        want
	}{
		{
			name:        "negative test #1 POST",
			method:      http.MethodPost,
			path:        "/",
			body:        bytes.NewReader([]byte("yandex.ru")),
			contentType: "biba",
			want: want{
				code:        415,
				response:    "Unsupported Media Type\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:        "positive test #1 POST",
			method:      http.MethodPost,
			path:        "/",
			body:        bytes.NewReader([]byte("yandex.ru")),
			contentType: "text/plain",
			want: want{
				code:        201,
				response:    templateURL,
				contentType: "text/plain",
			},
		},
		{
			name:        "positive test #1 GET",
			method:      http.MethodGet,
			path:        "/",
			body:        bytes.NewReader([]byte("yandex.ru")),
			contentType: "biba",
			want: want{
				code:        307,
				response:    "",
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.method == http.MethodGet {
				test.path = fmt.Sprintf("/%s", tmpURL)
			}
			request := httptest.NewRequest(
				test.method,
				test.path,
				test.body,
			)
			request.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()

			router := InitRoutes(app)

			router.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Contains(t, string(resBody), test.want.response)
			if test.method == http.MethodPost &&
				test.want.code == http.StatusCreated {
				if assert.Contains(t, string(resBody), templateURL) {
					u, err := url.Parse(string(resBody))
					require.NoError(t, err)
					tmpURL = path.Base(u.Path)
				}
			}
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func ExampleShortenerRoutes_сreateShortURL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage := repository.NewCache(context.Background(), "", tmpFile)
	shortenerService := &service.ShortenerService{
		Storage:  storage,
		DeleteCh: make(chan service.DeleteTask, 1),
	}
	router := ShortenerRoutes(shortenerService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		bytes.NewBufferString(`{"url":"https://practicum.yandex.ru"}`),
	)
	req.Header.Set("Content-Type", "text/plain")

	ctx := context.WithValue(
		req.Context(),
		service.UserIDContextKey(),
		"test-user",
	)

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)

	// Output:
	// 201
}

func ExampleShortenerRoutes_getOriginalURL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage := repository.NewCache(context.Background(), "", tmpFile)
	storage.URLCache = map[string]repository.URL{
		"abc123": {
			UUID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Original:    "https://practicum.yandex.ru",
			Short:       "abc123",
			UserID:      "test-user",
			DeletedFlag: false,
		},
	}

	shortenerService := &service.ShortenerService{
		Storage:  storage,
		DeleteCh: make(chan service.DeleteTask, 1),
	}
	router := ShortenerRoutes(shortenerService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc123",
		bytes.NewBufferString(""),
	)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(
		req.Context(),
		service.UserIDContextKey(),
		"test-user",
	)

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)

	// Output:
	// 307
}

func ExampleShortenerRoutes_сreateShortURLApi(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage := repository.NewCache(context.Background(), "", tmpFile)
	shortenerService := &service.ShortenerService{
		Storage:  storage,
		DeleteCh: make(chan service.DeleteTask, 1),
	}
	router := ShortenerRoutes(shortenerService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		bytes.NewBufferString(`{"url":"https://practicum.yandex.ru"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(
		req.Context(),
		service.UserIDContextKey(),
		"test-user",
	)

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)

	// Output:
	// 201
}

func ExampleShortenerRoutes_createShortURLBatchApi(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage := repository.NewCache(context.Background(), "", tmpFile)
	shortenerService := &service.ShortenerService{
		Storage:  storage,
		DeleteCh: make(chan service.DeleteTask, 1),
	}
	router := ShortenerRoutes(shortenerService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten/batch",
		bytes.NewBufferString(`[
        {
            "correlation_id":"1",
            "original_url":"https://practicum.yandex.ru"
        }
    ]`),
	)

	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(
		req.Context(),
		service.UserIDContextKey(),
		"test-user",
	)

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)

	// Output:
	// 201
}

func ExampleShortenerRoutes_getUsersURLs(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage := repository.NewCache(context.Background(), "", tmpFile)
	storage.URLCache = map[string]repository.URL{
		"abc123": {
			UUID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Original:    "https://practicum.yandex.ru",
			Short:       "abc123",
			UserID:      "test-user",
			DeletedFlag: false,
		},
	}
	shortenerService := &service.ShortenerService{
		Storage:  storage,
		DeleteCh: make(chan service.DeleteTask, 1),
	}
	router := ShortenerRoutes(shortenerService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/user/urls",
		bytes.NewBufferString(`{"url":"https://practicum.yandex.ru"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(
		req.Context(),
		service.UserIDContextKey(),
		"test-user",
	)

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)

	// Output:
	// 200
}

func ExampleShortenerRoutes_deleteUsersURLs(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	storage := repository.NewCache(context.Background(), "", tmpFile)
	storage.URLCache = map[string]repository.URL{
		"abc123": {
			UUID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Original:    "https://practicum.yandex.ru",
			Short:       "abc123",
			UserID:      "test-user",
			DeletedFlag: false,
		},
	}
	shortenerService := &service.ShortenerService{
		Storage:  storage,
		DeleteCh: make(chan service.DeleteTask, 1),
	}
	router := ShortenerRoutes(shortenerService)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/user/urls",
		bytes.NewBufferString(`["abc123"]`),
	)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(
		req.Context(),
		service.UserIDContextKey(),
		"test-user",
	)

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)

	// Output:
	// 202
}

func TestApp_registerDispatcher(t *testing.T) {
	app := &App{
		Config: config.Config{
			FileLogging: config.LoggerInfo{
				Enable: true,
				Path:   "/tmp/test.log",
			},
			RemoteLogging: config.LoggerInfo{
				Enable: true,
				Path:   "http://localhost:8080/log",
			},
		},
	}

	d, err := logs.NewDispatcher(app.Config)

	if d == nil {
		t.Fatal("dispatcher is nil")
	}
	if err != nil {
		t.Fatal("error is not nil")
	}
}

func TestApp_validateApp_nilLogger(t *testing.T) {
	app := &App{}

	app.validateApp()

	if app.Logger == nil {
		t.Fatal("logger should be initialized")
	}
}
