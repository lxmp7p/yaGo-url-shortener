package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/logger"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Service *service.ShortenerService
}

func NewHandler(s *service.ShortenerService) *Handler {
	return &Handler{
		Service: s,
	}
}

type App struct {
	Config   config.Config
	Storage  service.URLstorage
	Logger   *logrus.Logger
	Database *sql.DB
}

func InitRoutes(app App) chi.Router {
	app.validateApp()
	shortenerService := &service.ShortenerService{
		Config:     app.Config,
		Storage:    app.Storage,
		Database:   app.Database,
		DeleteCh:   make(chan service.DeleteTask, 1),
		Dispatcher: app.registerDispatcher(),
	}
	shortenerService.StartDeleteWorker()

	handler := NewHandler(shortenerService)

	apiRouter := chi.NewRouter()
	apiRouter.Use(CompressMiddleware())
	apiRouter.Use(LoggingMiddleware(app.Logger))
	apiRouter.Use(handler.AuthMiddleware)

	apiRouter.Mount("/", ShortenerRoutes(shortenerService))
	return apiRouter
}

func (app *App) validateApp() {
	if app.Logger == nil {
		app.Logger = logrus.New()
	}
}

func (app *App) registerDispatcher() *logger.Dispatcher {
	dispatcher := &logger.Dispatcher{}

	if app.Config.FileLogging.Enable {
		dispatcher.Register(&logger.FileObserver{Path: app.Config.FileLogging.Path})
	}

	if app.Config.RemoteLogging.Enable {
		dispatcher.Register(&logger.HTTPObserver{
			URL:    app.Config.RemoteLogging.Path,
			Client: &http.Client{Timeout: 2 * time.Second},
		})
	}

	return dispatcher
}
