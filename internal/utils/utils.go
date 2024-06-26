package utils

import (
	"database/sql"
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
	BatchUpdate(metrics []metrics.Metric)
}

func NewRouter(storage metricRepository, key string, db *sql.DB, subnet string) chi.Router {
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

	router.Group(func(r chi.Router) {
		r.Use(middleware.SubnetCheck(subnet))
		r.Route("/update", func(ru chi.Router) {
			ru.Post("/", handlers.JSONUpdateHandler(storage, key))
			ru.Post("/{kind}/{name}/{value}", handlers.UpdateStorageHandler(storage, key))
		})
		r.Post("/updates/", handlers.MetricsUpdateHandler(storage))
	})

	router.Get("/ping", handlers.PingDatabaseHandler(db))

	router.Mount("/debug", chiMiddleware.Profiler())

	return router
}

func UpdateDurVar(envName string, fl *time.Duration, configValue time.Duration) time.Duration {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return *fl
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return *fl
	}

	if result == 0 && *fl == 0 {
		return configValue
	}

	return time.Duration(result) * time.Second
}

const (
	DefaultAddress = "127.0.0.1:8080"
)

func UpdateStringVar(envName string, fl *string, configValue string) string {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return *fl
	}

	if value == "" && *fl == "" {
		return configValue
	}

	return value
}

func UpdateBoolVar(envName string, fl *bool, configValue bool) bool {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return *fl
	}

	if value == "" && !*fl {
		return configValue
	}

	return value == "true"
}

func UpdateIntVar(envName string, fl *int, configValue int) int {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return *fl
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return *fl
	}

	if result == 0 && *fl == 0 {
		return configValue
	}

	return result
}
