package utils

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/handlers"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"os"
	"strconv"
	"time"
)

type metricRepository interface {
	GetMetricsMap() map[string]metrics.Metric
	GetMetric(name string) (metrics.Metric, error)
	Update(metrics.Metric)
}

func NewRouter(storage metricRepository, key, dbDsn string) chi.Router {
	router := chi.NewRouter()
	router.Use(
		chiMiddleware.RequestID,
		chiMiddleware.RealIP,
		chiMiddleware.Logger,
		chiMiddleware.Recoverer,
		middleware.Compress,
		middleware.Decompress,
	)

	router.Get("/", handlers.PrintStorageHandler(storage))

	router.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.JSONPrintHandler(storage, key))
		r.Get("/{kind}/{name}", handlers.PrintValueHandler(storage, key))
	})

	router.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.JSONUpdateHandler(storage, key))
		r.Post("/{kind}/{name}/{value}", handlers.UpdateStorageHandler(storage, key))
	})

	router.Get("/ping", handlers.PingDatabaseHandler(dbDsn))

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
