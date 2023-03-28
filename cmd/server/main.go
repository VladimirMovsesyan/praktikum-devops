package main

import (
	"context"
	"flag"
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
	defaultStore     = 300 * time.Second
	defaultStoreFile = "/tmp/devops-metrics-db.json"
	defaultRestore   = true
)

var (
	flAddr          *string
	flStoreInterval *time.Duration
	flStoreFile     *string
	flRestore       *bool
)

func init() {
	log.Println("server init...")
	flAddr = flag.String("a", utils.DefaultAddress, "Server IP address")           // ADDRESS
	flStoreInterval = flag.Duration("i", defaultStore, "Interval of storing data") // STORE_INTERVAL
	flStoreFile = flag.String("f", defaultStoreFile, "Path to storage file")       // STORE_FILE
	flRestore = flag.Bool("r", defaultRestore, "Is need to restore storage")       // RESTORE
	flag.Parse()
}

func main() {
	storage := repository.NewMemStorage()
	router := utils.NewRouter(storage)
	address := utils.UpdateStringVar(
		"ADDRESS",
		flAddr,
	)
	server := http.Server{Addr: address, Handler: router}

	storeInterval := time.NewTicker(
		utils.UpdateDurVar(
			"STORE_INTERVAL",
			flStoreInterval,
		),
	)

	storeFilePath := utils.UpdateStringVar(
		"STORE_FILE",
		flStoreFile,
	)

	restore := utils.UpdateBoolVar(
		"RESTORE",
		flRestore,
	)

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
