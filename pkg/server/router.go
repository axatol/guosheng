package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(addr string) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.RequestLogger(requestLogFormatter{}))
	router.Use(middleware.Recoverer)
	router.Use(middlewareContentType("application/json"))
	router.Use(middleware.AllowContentType("application/json", "text/json"))
	return router
}
