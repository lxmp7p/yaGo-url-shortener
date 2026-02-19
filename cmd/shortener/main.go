package main

import (
	"net/http"

	"github.com/lxmp7p/yaGo-url-shortener/internal/handler"
)

func main() {
	http.HandleFunc(`/`, handler.Handler)

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
