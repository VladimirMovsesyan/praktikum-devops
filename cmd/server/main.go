package main

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/handlers"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
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
	storage := &repository.MemStorage{}
	http.HandleFunc("/update/", handlers.UpdateStorageHandler(storage))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
