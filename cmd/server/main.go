package main

import (
	"context"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/cache"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultStore     = 300
	defaultStoreFile = "/tmp/devops-metrics-db.json"
	defaultRestore   = true
)

func main() {
	storage := repository.NewMemStorage()
	router := utils.NewRouter(storage)
	address := utils.UpdateStringEnv("ADDRESS", utils.DefaultAddress)
	server := http.Server{Addr: address, Handler: router}

	storeInterval := time.NewTicker(
		time.Duration(
			utils.UpdateIntEnv(
				"STORE_INTERVAL",
				defaultStore,
			),
		) * time.Second,
	)
	storeFilePath := utils.UpdateStringEnv("STORE_FILE", defaultStoreFile)
	restore := utils.UpdateBoolEnv("RESTORE", defaultRestore)

	if restore {
		err := cache.ImportData(storeFilePath, storage)
		if err != nil {
			log.Println(err)
			return
		}
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("HTTP server ListenAndServe:", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	for {
		select {
		case sig := <-signals:
			log.Println(sig.String())

			if err := server.Shutdown(context.Background()); err != nil {
				log.Println("HTTP server Shutdown:", err)
			}

			err := cache.ExportData(storeFilePath, storage)
			if err != nil {
				log.Println(err)
				return
			}

			return
		case <-storeInterval.C:
			err := cache.ExportData(storeFilePath, storage)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
