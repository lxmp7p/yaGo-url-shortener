package service

import (
	"log/slog"
	"net/http"
)

func (s *ShortenerService) PingDatabase(res http.ResponseWriter, req *http.Request) {
	if err := s.Database.Ping(); err != nil {
		slog.Error(err.Error())
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
