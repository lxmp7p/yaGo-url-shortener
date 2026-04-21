package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config/db"
	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lxmp7p/yaGo-url-shortener/internal/handler"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.NewConfig()
	cfg.InitConfig()

	logger := logrus.New()

	database, err := db.InitDB(cfg, logger)

	if err != nil {
		logger.Fatal(err)
	}

	r := handler.InitRoutes(handler.App{
		Config:   cfg,
		Storage:  selectStorage(cfg, database),
		Logger:   logger,
		Database: database,
	})

	logger.Infof("Starting server on %s", cfg.Addr)
	err = http.ListenAndServe(cfg.Addr, r)
	logger.Fatalf("Server stopped: %v", err)

}

func selectStorage(cfg config.Config, database *sql.DB) service.URLstorage {
	if cfg.DatabaseDsn != "" {
		return repository.NewDatabaseCache(database)
	}
	if cfg.FileStoragePath != "" {
		return repository.NewCache(context.Background(), cfg.FileStoragePath)
	}
	return repository.NewCache(context.Background(), "")
}
