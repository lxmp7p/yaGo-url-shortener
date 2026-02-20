package main

import (
	"net/http"

	"github.com/lxmp7p/yaGo-url-shortener/internal/handler"
)

func main() {
	r := handler.InitRoutes()
	http.ListenAndServe(":8080", r)
}
