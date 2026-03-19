package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
	"github.com/sirupsen/logrus"
)

type App struct {
	Config  config.Config
	Storage service.URLstorage
	Logger  *logrus.Logger
}

func InitRoutes(app App) chi.Router {
	app.validateApp()

	shortenerService := &service.ShortenerService{
		Config:  app.Config,
		Storage: app.Storage,
	}
	apiRouter := chi.NewRouter()
	//apiRouter.Use(CompressMiddleware())
	apiRouter.Use(LoggingMiddleware(app.Logger))

	apiRouter.Mount("/", ShortenerRoutes(shortenerService))
	return apiRouter
}

func (app App) validateApp() {
	if app.Logger == nil {
		app.Logger = logrus.New()
	}
}
