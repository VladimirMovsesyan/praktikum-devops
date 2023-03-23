package utils

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(storage handlers.MetricRepository) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)

	router.Get("/", handlers.PrintStorageHandler(storage))

	router.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.JSONPrintHandler(storage))
		r.Get("/{kind}/{name}", handlers.PrintValueHandler(storage))
	})

	router.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.JSONUpdateHandler(storage))
		r.Post("/{kind}/{name}/{value}", handlers.UpdateStorageHandler(storage))
	})

	return router
}
