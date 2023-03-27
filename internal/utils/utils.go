package utils

import (
	"flag"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"os"
	"strconv"
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

func updateIntFlag(flagName string, defaultValue int64, usage string) int64 {
	fl := flag.Int64(flagName, defaultValue, usage)
	flag.Parse()
	return *fl
}

func UpdateIntVar(envName, flagName string, defaultValue int64, usage string) int64 {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return updateIntFlag(flagName, defaultValue, usage)
	}

	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return updateIntFlag(flagName, defaultValue, usage)
	}

	return result
}

const (
	DefaultAddress = "127.0.0.1:8080"
)

func updateStringFlag(flagName, defaultValue, usage string) string {
	fl := flag.String(flagName, defaultValue, usage)
	flag.Parse()
	return *fl
}

func UpdateStringVar(envName, flagName, defaultValue, usage string) string {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return updateStringFlag(flagName, defaultValue, usage)
	}
	return value
}

func updateBoolFlag(flagName string, defaultValue bool, usage string) bool {
	fl := flag.Bool(flagName, defaultValue, usage)
	flag.Parse()
	return *fl
}

func UpdateBoolVar(envName, flagName string, defaultValue bool, usage string) bool {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return updateBoolFlag(flagName, defaultValue, usage)
	}
	return value == "true"
}
