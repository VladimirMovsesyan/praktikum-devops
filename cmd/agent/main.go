package main

import (
	"flag"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/clients"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/config"
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
	flConfig     *bool          // CONFIG
)

func parseFlags() {
	log.Println("agent init...")
	flAddr = flag.String("a", utils.DefaultAddress, "Server IP address")          // ADDRESS
	flPoll = flag.Duration("p", defaultPoll, "Interval of polling metrics")       // POLL_INTERVAL
	flReport = flag.Duration("r", defaultReport, "Interval of reporting metrics") // REPORT_INTERVAL
	flKey = flag.String("k", "", "Hash key")                                      // KEY
	flLimit = flag.Int("l", defaultLimit, "Limit of requests rate")               // RATE_LIMIT
	flCrypto = flag.String("crypto-key", "", "Path to public crypto key")         // CRYPTO_KEY
	flConfig = flag.Bool("config", false, "Configuration by config json file")    // CONFIG
	flag.Parse()
}

func main() {
	log.Println("Build version:", buildVersion)
	log.Println("Build date:", buildDate)
	log.Println("Build commit:", buildCommit)
	parseFlags()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	conf := utils.UpdateBoolVar(
		"CONFIG",
		flConfig,
		false,
	)

	configuration := &config.AgentConfig{}

	if conf {
		var err error
		configuration, err = config.NewAgentConfig()
		if err != nil {
			log.Println(err)
			return
		}
	}

	address := utils.UpdateStringVar(
		"ADDRESS",
		flAddr,
		configuration.Address,
	)

	cPoll := defaultPoll
	if conf {
		var err error
		cPoll, err = time.ParseDuration(configuration.PollInterval)
		if err != nil {
			log.Println(err)
			return
		}
	}

	cReport := defaultReport
	if conf {
		var err error
		cReport, err = time.ParseDuration(configuration.ReportInterval)
		if err != nil {
			log.Println(err)
			return
		}
	}

	// Creating poll and report intervals
	pollInterval := time.NewTicker(
		utils.UpdateDurVar(
			"POLL_INTERVAL",
			flPoll,
			cPoll,
		),
	)
	reportInterval := time.NewTicker(
		utils.UpdateDurVar(
			"REPORT_INTERVAL",
			flReport,
			cReport,
		),
	)

	key := utils.UpdateStringVar(
		"KEY",
		flKey,
		configuration.Key,
	)

	limit := utils.UpdateIntVar(
		"RATE_LIMIT",
		flLimit,
		configuration.Limit,
	)

	keyPath := utils.UpdateStringVar(
		"CRYPTO_KEY",
		flCrypto,
		configuration.Crypto,
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
