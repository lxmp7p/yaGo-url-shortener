package handler

import (
	"github.com/go-chi/chi/v5"
)

func InitRoutes() chi.Router {
	apiRouter := chi.NewRouter()
	apiRouter.Mount("/", Routes())
	return apiRouter
}
