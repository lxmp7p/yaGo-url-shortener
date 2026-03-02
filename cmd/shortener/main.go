package main

import (
	"net/http"

	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/handler"
)

func main() {
	cfg := config.InitConfig()

	r := handler.InitRoutes(handler.App{
		Config: cfg,
	})
	http.ListenAndServe(cfg.Addr, r)
}
