package main

import (
	"net/http"

	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config/db"

	"github.com/lxmp7p/yaGo-url-shortener/internal/handler"
	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.NewConfig()
	cfg.InitConfig()

	logger := logrus.New()

	db := db.InitDB(cfg)

	r := handler.InitRoutes(handler.App{
		Config:  cfg,
		Storage: repository.NewCache(cfg.FileStoragePath),
		Logger:  logger,
		Database: db,
	})
	http.ListenAndServe(cfg.Addr, r)
}
