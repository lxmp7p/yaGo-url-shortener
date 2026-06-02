package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config/db"
	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"

	_ "net/http/pprof"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lxmp7p/yaGo-url-shortener/internal/handler"
	"github.com/sirupsen/logrus"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	cfg := config.NewConfig()
	cfg.InitConfig()

	logger := logrus.New()

	database, err := db.InitDB(cfg, logger)

	if err != nil {
		logger.Fatal(err)
	}

	storage, err := selectStorage(cfg, database)
	if err != nil {
		logger.Fatal(err)
	}

	r := handler.InitRoutes(handler.App{
		Config:   cfg,
		Storage:  storage,
		Logger:   logger,
		Database: database,
	})

	logger.Infof("Starting server on %s", cfg.Addr)
	err = http.ListenAndServe(cfg.Addr, r)
	logger.Fatalf("Server stopped: %v", err)

}

func selectStorage(cfg config.Config, database *sql.DB) (service.URLstorage, error) {
	file, err := os.OpenFile(cfg.FileStoragePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	if cfg.DatabaseDsn != "" {
		return repository.NewDatabaseCache(database), nil
	}
	if cfg.FileStoragePath != "" {
		return repository.NewCache(context.Background(), cfg.FileStoragePath, file), nil
	}
	return repository.NewCache(context.Background(), "", file), nil
}
