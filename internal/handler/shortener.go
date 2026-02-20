package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
)

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{short_url}", service.GetOriginalURL)
	r.Post("/", service.GetShortURL)

	return r
}
