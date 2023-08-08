package main

import (
	"flag"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/clients"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultPoll   = 2 * time.Second
	defaultReport = 10 * time.Second
	defaultLimit  = 1
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
	flAddr       *string        // ADDRESS
	flPoll       *time.Duration // POLL_INTERVAL
	flReport     *time.Duration // REPORT_INTERVAL
	flKey        *string        // KEY
	flLimit      *int           // RATE_LIMIT
	flCrypto     *string        // CRYPTO_KEY
)

func parseFlags() {
	log.Println("agent init...")
	flAddr = flag.String("a", utils.DefaultAddress, "Server IP address")          // ADDRESS
	flPoll = flag.Duration("p", defaultPoll, "Interval of polling metrics")       // POLL_INTERVAL
	flReport = flag.Duration("r", defaultReport, "Interval of reporting metrics") // REPORT_INTERVAL
	flKey = flag.String("k", "", "Hash key")                                      // KEY
	flLimit = flag.Int("l", defaultLimit, "Limit of requests rate")               // RATE_LIMIT
	flCrypto = flag.String("crypto-key", "", "Path to public crypto key")         // CRYPTO_KEY
	flag.Parse()
}

func main() {
	log.Println("Build version:", buildVersion)
	log.Println("Build date:", buildDate)
	log.Println("Build commit:", buildCommit)
	parseFlags()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	address := utils.UpdateStringVar(
		"ADDRESS",
		flAddr,
	)

	// Creating poll and report intervals
	pollInterval := time.NewTicker(
		utils.UpdateDurVar(
			"POLL_INTERVAL",
			flPoll,
		),
	)
	reportInterval := time.NewTicker(
		utils.UpdateDurVar(
			"REPORT_INTERVAL",
			flReport,
		),
	)

	key := utils.UpdateStringVar(
		"KEY",
		flKey,
	)

	limit := utils.UpdateIntVar(
		"RATE_LIMIT",
		flLimit,
	)

	keyPath := utils.UpdateStringVar(
		"CRYPTO_KEY",
		flCrypto,
	)

	// Creating worker pool
	wp := clients.NewWorkerPool(limit, address, key, keyPath)

	// Worker pool process start
	wp.Run()

	// Agent's process
	for {
		select {
		case <-pollInterval.C:
			// Updating metrics
			wp.AddTask("updateMem")
			wp.AddTask("updateGopsutil")
		case <-reportInterval.C:
			// Sending metrics
			wp.AddTask("upload")
		case sig := <-signals:
			wp.Stop()
			log.Println("Got signal:", sig.String())
			return
		}
	}
}
