package main

import (
	"net/http"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
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
	if cfg.DatabaseDsn != "" {
		runMigrations(cfg.DatabaseDsn, logger)
	}

	r := handler.InitRoutes(handler.App{
		Config:  cfg,
		Storage: selectStorage(cfg),
		Logger:  logger,
	})
	http.ListenAndServe(cfg.Addr, r)
}

func selectStorage(cfg config.Config) service.URLstorage {
	if cfg.DatabaseDsn != "" {
		return repository.NewDatabaseCache(db.InitDB(cfg))
	}
	if cfg.FileStoragePath != "" {
		return repository.NewCache(cfg.FileStoragePath)
	}
	return nil
}

func runMigrations(DSN string, logger *logrus.Logger) {
	path, err := filepath.Abs("../../migrations")
	if err != nil {
		logger.Fatal("Migration not found")
	}
	migrationsPath := "file://" + path
	dbURL := DSN
	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		logger.Fatalf("Failed to initialize migrate: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatalf("Migration failed: %v", err)
	}

	logger.Println("Migrations applied successfully!")
}
