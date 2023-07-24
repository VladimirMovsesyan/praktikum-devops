package main

import (
	"context"
	"database/sql"
	"flag"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/cache"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/config"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/crypt"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
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

type metricRepository interface {
	GetMetricsMap() map[string]metrics.Metric
	GetMetric(name string) (metrics.Metric, error)
	Update(metrics.Metric)
	BatchUpdate(metrics []metrics.Metric)
}

var (
	buildVersion    = "N/A"
	buildDate       = "N/A"
	buildCommit     = "N/A"
	flAddr          *string        // ADDRESS
	flStoreInterval *time.Duration // STORE_INTERVAL
	flStoreFile     *string        // STORE_FILE
	flRestore       *bool          // RESTORE
	flKey           *string        // KEY
	flDSN           *string        // DATABASE_DSN
	flCrypt         *string        // CRYPTO_KEY
	flConfig        *bool          // CONFIG
)

func parseFlags() {
	log.Println("server init...")
	flAddr = flag.String("a", utils.DefaultAddress, "Server IP address")           // ADDRESS
	flStoreInterval = flag.Duration("i", defaultStore, "Interval of storing data") // STORE_INTERVAL
	flStoreFile = flag.String("f", defaultStoreFile, "Path to storage file")       // STORE_FILE
	flRestore = flag.Bool("r", defaultRestore, "Is need to restore storage")       // RESTORE
	flKey = flag.String("k", "", "Hash key")                                       // KEY
	flDSN = flag.String("d", "", "Data source name")                               // DATABASE_DSN
	flCrypt = flag.String("crypto-key", "", "Path to private crypto key")          // CRYPTO_KEY
	flConfig = flag.Bool("config", false, "Configuration by config json file")     // CONFIG
	flag.Parse()
}

func main() {
	log.Println("Build version:", buildVersion)
	log.Println("Build date:", buildDate)
	log.Println("Build commit:", buildCommit)
	parseFlags()

	var configuration *config.ServerConfig

	conf := utils.UpdateBoolVar(
		"CONFIG",
		flConfig,
		false,
	)
	if conf {
		var err error
		configuration, err = config.NewServerConfig()
		if err != nil {
			log.Println(err)
			return
		}
	}

	log.Println(configuration)

	key := utils.UpdateStringVar(
		"KEY",
		flKey,
		configuration.Key,
	)
	dbDSN := utils.UpdateStringVar(
		"DATABASE_DSN",
		flDSN,
		configuration.Dsn,
	)

	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var storage metricRepository
	switch dbDSN {
	case "":
		storage = repository.NewMemStorage()
	default:
		storage = repository.NewPostgreStorage(db)
	}

	router := utils.NewRouter(storage, key, db)

	cryptoPath := utils.UpdateStringVar(
		"CRYPTO_KEY",
		flCrypt,
		configuration.Crypt,
	)
	if cryptoPath != "" {
		c, err := crypt.New(crypt.WithPrivateKey(cryptoPath))
		if err != nil {
			log.Fatal(err)
		}
		router.Use(c.GetDecryptMiddleware())
	}

	address := utils.UpdateStringVar(
		"ADDRESS",
		flAddr,
		configuration.Address,
	)
	server := http.Server{Addr: address, Handler: router}

	cTime := defaultStore
	if conf {
		cTime, err = time.ParseDuration(configuration.StoreInterval)
		if err != nil {
			log.Println(err)
			return
		}
	}

	storeInterval := time.NewTicker(
		utils.UpdateDurVar(
			"STORE_INTERVAL",
			flStoreInterval,
			cTime,
		),
	)

	storeFilePath := utils.UpdateStringVar(
		"STORE_FILE",
		flStoreFile,
		configuration.StoreFile,
	)

	restore := utils.UpdateBoolVar(
		"RESTORE",
		flRestore,
		configuration.Restore,
	)

	if restore && dbDSN == "" {
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

			if dbDSN == "" {
				log.Println("exporting data after shutdown")
				err := cache.ExportData(storeFilePath, storage)
				if err != nil {
					log.Println(err)
				}
			}

			return
		case <-storeInterval.C:
			if dbDSN == "" {
				log.Println("normal exporting data")
				err := cache.ExportData(storeFilePath, storage)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
