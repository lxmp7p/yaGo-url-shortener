package main

import (
	"net/http"

	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/handler"
	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
)

func main() {
	cfg := config.NewConfig()
	cfg.InitConfig()

	r := handler.InitRoutes(handler.App{
		Config:  cfg,
		Storage: repository.NewCache(),
	})
	http.ListenAndServe(cfg.Addr, r)
}
