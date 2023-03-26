package utils

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"os"
	"strconv"
	"time"
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

func UpdateInterval(envName string, defaultValue time.Duration) time.Duration {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}

	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return time.Duration(result)
}

const (
	DefaultAddress = "127.0.0.1:8080"
)

func UpdateAddress(envName string, defaultValue string) string {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}
	return value
}
