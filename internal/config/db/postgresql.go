package db

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/sirupsen/logrus"
)

func InitDB(config config.Config, logger *logrus.Logger) (*sql.DB, error) {
	if config.DatabaseDsn != "" {
		logger.Info("Running migrations...")
		runMigrations(config.DatabaseDsn, logger)
		logger.Info("Migrations done")
	}

	db, err := sql.Open("pgx", config.DatabaseDsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(DSN string, logger *logrus.Logger) {
	migrationsPath := "file://migrations"
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
