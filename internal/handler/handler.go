package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
)

type App struct {
	Config  config.Config
	Storage service.URLstorage
}

func InitRoutes(app App) chi.Router {
	shortenerService := &service.ShortenerService{
		Config:  app.Config,
		Storage: app.Storage,
	}
	apiRouter := chi.NewRouter()
	apiRouter.Mount("/", ShortenerRoutes(shortenerService))
	return apiRouter
}
