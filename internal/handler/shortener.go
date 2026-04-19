package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
)

func ShortenerRoutes(shortenerService *service.ShortenerService) chi.Router {
	r := chi.NewRouter()
	r.Post("/", shortenerService.GetShortURL)
	r.Get("/{short_url}", shortenerService.GetOriginalURL)

	r.Post("/api/shorten", shortenerService.GetShortURLApi)
	r.Post("/api/shorten/batch", shortenerService.GetShortURLBatchAPI)

	r.Get("/ping", shortenerService.PingDatabase)

	return r
}
