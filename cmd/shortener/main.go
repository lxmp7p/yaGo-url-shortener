package main

import (
	"net/http"

	"yaGo-url-shortener/internal/handler"
)

func main() {
	http.HandleFunc(`/`, handler.Handler)

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
