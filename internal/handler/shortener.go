package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
)

// Регистрирует все роуты приложения
func ShortenerRoutes(shortenerService *service.ShortenerService) chi.Router {
	r := chi.NewRouter()
	r.Post("/", shortenerService.CreateShortURL)
	r.Get("/{short_url}", shortenerService.GetOriginalURL)

	r.Post("/api/shorten", shortenerService.CreateShortURLApi)
	r.Post("/api/shorten/batch", shortenerService.CreateShortURLBatchAPI)
	r.Get("/api/user/urls", shortenerService.GetUsersURLs)
	r.Delete("/api/user/urls", shortenerService.DeleteUsersURLs)

	r.With(TrustedSubnetMiddleware(shortenerService.Config.TrustedSubnet)).
		Get("/api/internal/stats", shortenerService.GetStats)

	r.Get("/ping", shortenerService.PingDatabase)
	return r
}
