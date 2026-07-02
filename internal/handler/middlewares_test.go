package handler

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
	"github.com/sirupsen/logrus"
)

func TestCompressResponseWriter_GzipEnabled(t *testing.T) {

	resp := httptest.NewRecorder()

	w := &compressResponseWriter{
		ResponseWriter: resp,
	}

	w.Header().Set(ContentType, "text/html")

	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	if n == 0 {
		t.Fatal("expected bytes written")
	}

	if resp.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected gzip header")
	}
}

func TestApp_validateApp_setsLoggerIfNil(t *testing.T) {
	app := &App{}

	app.validateApp()

	if app.Logger == nil {
		t.Fatal("expected logger to be initialized")
	}
}

type fakeResponseWriter struct {
	http.ResponseWriter
	body       []byte
	statusCode int
}

func (f *fakeResponseWriter) Write(b []byte) (int, error) {
	f.body = append(f.body, b...)
	return len(b), nil
}

func (f *fakeResponseWriter) WriteHeader(code int) {
	f.statusCode = code
}

func TestLoggingMiddleware_logsRequest(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	mw := LoggingMiddleware(logger)

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if !strings.Contains(buf.String(), "method:GET") {
		t.Fatal("log not written or incorrect")
	}
}

func TestCompressMiddleware_decompressesGzipBody(t *testing.T) {
	var out bytes.Buffer

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("hello"))
	gz.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set(ContentEncoding, Gzip)

	mw := CompressMiddleware()

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(&out, r.Body)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if out.String() != "hello" {
		t.Fatalf("expected hello, got %s", out.String())
	}
}

type mockService struct{}

func (m *mockService) Sign(userID string) string {
	return "sig"
}

func (m *mockService) Verify(userID, sig string) bool {
	return true
}

func TestAuthMiddleware_createsCookieIfMissing(t *testing.T) {
	svc := &service.ShortenerService{}

	h := &Handler{Service: svc}

	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.AuthMiddleware(next).ServeHTTP(rec, req)

	if !called {
		t.Fatal("next not called")
	}

	cookieHeader := rec.Header().Get("Set-Cookie")
	if cookieHeader == "" {
		t.Fatal("expected Set-Cookie header")
	}
}
