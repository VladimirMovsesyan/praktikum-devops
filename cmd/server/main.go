package main

import (
	"context"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/handlers"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	router := chi.NewRouter()
	storage := repository.NewMemStorage()
	router.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)

	router.Get("/", handlers.PrintStorageHandler(storage))
	router.Get("/value/{kind}/{name}", handlers.PrintValueHandler(storage))
	router.Post("/update/{kind}/{name}/{value}", handlers.UpdateStorageHandler(storage))

	server := http.Server{Addr: ":8080", Handler: router}
	idle := make(chan struct{})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("HTTP server ListenAndServe:", err)
		}
		<-idle
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-signals

	if err := server.Shutdown(context.Background()); err != nil {
		log.Println("HTTP server Shutdown:", err)
	}
	close(idle)
}
