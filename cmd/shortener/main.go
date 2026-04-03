package main

import (
	"database/sql"
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
		logger.Info("Running migrations...")
		runMigrations(cfg.DatabaseDsn, logger)
		logger.Info("Migrations done")
	}

	database := db.InitDB(cfg)

	r := handler.InitRoutes(handler.App{
		Config:   cfg,
		Storage:  selectStorage(cfg, database),
		Logger:   logger,
		Database: database,
	})
	logger.Infof("Starting server on %s", cfg.Addr)
	err := http.ListenAndServe(cfg.Addr, r)
	logger.Fatalf("Server stopped: %v", err)

}

func selectStorage(cfg config.Config, database *sql.DB) service.URLstorage {
	if cfg.DatabaseDsn != "" {
		return repository.NewDatabaseCache(database)
	}
	if cfg.FileStoragePath != "" {
		return repository.NewCache(cfg.FileStoragePath)
	}
	return repository.NewCache("")
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
		logger.Errorf("Failed to initialize migrate: %v", err)
		return
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Errorf("Migration failed: %v", err)
		return
	}

	logger.Println("Migrations applied successfully!")
}
