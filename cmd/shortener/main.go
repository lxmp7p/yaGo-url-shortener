package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config/db"
	logs "github.com/lxmp7p/yaGo-url-shortener/internal/logger"
	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"

	_ "net/http/pprof"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lxmp7p/yaGo-url-shortener/internal/handler"
	"github.com/sirupsen/logrus"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	printBuildInfo()
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
	defer storage.Close()

	dispatcher, err := logs.NewDispatcher(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer dispatcher.Close()

	r := handler.InitRoutes(handler.App{
		Config:     cfg,
		Storage:    storage,
		Logger:     logger,
		Database:   database,
		Dispatcher: dispatcher,
	})

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	go func() {
		logger.Infof("Starting server on %s", cfg.Addr)

		var err error
		if cfg.EnableHTTPS {
			err = srv.ListenAndServeTLS("cert.pem", "key.pem")
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	<-ctx.Done()

	logger.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("server shutdown error: %v", err)
	}

	if err := storage.Close(); err != nil {
		logger.Errorf("storage close error: %v", err)
	}

	logger.Info("Server stopped")
}

func selectStorage(cfg config.Config, database *sql.DB) (service.URLstorage, error) {
	if cfg.DatabaseDsn != "" {
		return repository.NewDatabaseCache(database), nil
	}

	file, err := os.OpenFile(cfg.FileStoragePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	if cfg.FileStoragePath != "" {
		return repository.NewCache(context.Background(), cfg.FileStoragePath, file), nil
	}
	return repository.NewCache(context.Background(), "", file), nil
}

func printBuildInfo() {
	log.Printf("Build version: %s\n", buildVersion)
	log.Printf("Build date: %s\n", buildDate)
	log.Printf("Build commit: %s\n", buildCommit)
}
