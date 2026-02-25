package main

import (
	"net/http"

	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/handler"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic("failed init config")
	}

	r := handler.InitRoutes(handler.App{
		Config: cfg,
	})
	http.ListenAndServe(cfg.Addr, r)
}
