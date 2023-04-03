package main

import (
	"context"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	storage := repository.NewMemStorage()
	router := utils.NewRouter(storage)
	address := utils.UpdateAddress("ADDRESS", utils.DefaultAddress)
	server := http.Server{Addr: address, Handler: router}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("HTTP server ListenAndServe:", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-signals
	log.Println(sig.String())

	if err := server.Shutdown(context.Background()); err != nil {
		log.Println("HTTP server Shutdown:", err)
	}
}
