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

func UpdateDurVar(envName string, fl *time.Duration) time.Duration {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return *fl
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return *fl
	}

	return time.Duration(result) * time.Second
}

const (
	DefaultAddress = "127.0.0.1:8080"
)

func UpdateStringVar(envName string, fl *string) string {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return *fl
	}
	return value
}

func UpdateBoolVar(envName string, fl *bool) bool {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return *fl
	}
	return value == "true"
}
