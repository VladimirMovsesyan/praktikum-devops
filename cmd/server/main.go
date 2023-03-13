package main

import (
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
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		sig := <-c
		log.Fatal(sig.String())
	}()
	router := chi.NewRouter()
	storage := &repository.MemStorage{}
	router.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)

	router.Get("/", handlers.PrintStorageHandler(storage))
	router.Get("/value/{kind}/{name}", handlers.PrintValueHandler(storage))
	router.Post("/update/{kind}/{name}/{value}", handlers.UpdateStorageHandler(storage))

	log.Fatal(http.ListenAndServe(":8080", router))
}
