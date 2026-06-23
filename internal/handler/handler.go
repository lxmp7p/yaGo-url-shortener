package handler

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/logger"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
	"github.com/sirupsen/logrus"
)

// Структура хэндлера хранящая в себе все необходимые сервисы
// generate:reset
type Handler struct {
	Service *service.ShortenerService
}

func NewHandler(s *service.ShortenerService) *Handler {
	return &Handler{
		Service: s,
	}
}

// Структура приложения хранящая все необходимые ресурсы, для запуска сервиса
// generate:reset
type App struct {
	Config     config.Config
	Storage    service.URLstorage
	Logger     *logrus.Logger
	Database   *sql.DB
	Dispatcher *logger.Dispatcher
}

// Создает сервис shortenerService, запускает воркер для удаления,
// создает хэндлер и подключает мидлвари к роуту
func InitRoutes(app App) chi.Router {
	app.validateApp()

	shortenerService := &service.ShortenerService{
		Config:     app.Config,
		Storage:    app.Storage,
		Database:   app.Database,
		DeleteCh:   make(chan service.DeleteTask, 1),
		Dispatcher: app.Dispatcher,
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

// Проверяет наличие логгера в приложении
func (app *App) validateApp() {
	if app.Logger == nil {
		app.Logger = logrus.New()
	}
}
