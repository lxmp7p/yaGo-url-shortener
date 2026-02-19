package handler

import (
	"net/http"

	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
)

func Handler(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		service.GetOriginalUrl(res, req)
		return
	}
	if req.Method == http.MethodPost {
		service.GetShortUrl(res, req)
		return
	}

	http.Error(res, "method not allowed", http.StatusMethodNotAllowed)
}
