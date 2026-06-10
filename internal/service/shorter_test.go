package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
