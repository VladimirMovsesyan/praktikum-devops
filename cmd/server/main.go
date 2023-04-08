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
	flAddr          *string        // ADDRESS
	flStoreInterval *time.Duration // STORE_INTERVAL
	flStoreFile     *string        // STORE_FILE
	flRestore       *bool          // RESTORE
	flKey           *string        // KEY
	flDbDSN         *string        // DATABASE_DSN
)

func parseFlags() {
	log.Println("server init...")
	flAddr = flag.String("a", utils.DefaultAddress, "Server IP address")           // ADDRESS
	flStoreInterval = flag.Duration("i", defaultStore, "Interval of storing data") // STORE_INTERVAL
	flStoreFile = flag.String("f", defaultStoreFile, "Path to storage file")       // STORE_FILE
	flRestore = flag.Bool("r", defaultRestore, "Is need to restore storage")       // RESTORE
	flKey = flag.String("k", "", "Hash key")                                       // KEY
	flDbDSN = flag.String("d", "", "Data source name")                             // DATABASE_DSN
	flag.Parse()
}

func main() {
	parseFlags()
	storage := repository.NewMemStorage()

	key := utils.UpdateStringVar(
		"KEY",
		flKey,
	)

	dbDsn := utils.UpdateStringVar(
		"DATABASE_DSN",
		flDbDSN,
	)

	router := utils.NewRouter(storage, key, dbDsn)
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
		log.Println("restoring data from", storeFilePath)
		err := cache.ImportData(storeFilePath, storage)
		if err != nil {
			log.Println(err)
			return
		}
	}

	go func() {
		log.Println("Listening:", address)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("HTTP server ListenAndServe:", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	for {
		select {
		case sig := <-signals:
			log.Println("Got signal:", sig.String())

			if err := server.Shutdown(context.Background()); err != nil {
				log.Println("HTTP server Shutdown:", err)
			}

			log.Println("exporting data after shutdown")
			err := cache.ExportData(storeFilePath, storage)
			if err != nil {
				log.Println(err)
			}

			return
		case <-storeInterval.C:
			log.Println("normal exporting data")
			err := cache.ExportData(storeFilePath, storage)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
