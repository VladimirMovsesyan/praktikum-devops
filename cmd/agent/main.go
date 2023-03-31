package main

import (
	"flag"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/clients"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
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
)

var (
	flAddr   *string        // ADDRESS
	flPoll   *time.Duration // POLL_INTERVAL
	flReport *time.Duration // REPORT_INTERVAL
)

func parseFlags() {
	log.Println("agent init...")
	flAddr = flag.String("a", utils.DefaultAddress, "Server IP address")          // ADDRESS
	flPoll = flag.Duration("p", defaultPoll, "Interval of polling metrics")       // POLL_INTERVAL
	flReport = flag.Duration("r", defaultReport, "Interval of reporting metrics") // REPORT_INTERVAL
	flag.Parse()
}

func main() {
	parseFlags()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Creating objects of metrics and client
	mtrcs := metrics.NewMetrics()

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

	// Agent's process
	for {
		select {
		case <-pollInterval.C:
			//Updating metrics
			go metrics.UpdateMetrics(mtrcs)
		case <-reportInterval.C:
			//Sending metrics
			go clients.MetricsUpload(mtrcs, flAddr)
		case sig := <-signals:
			log.Println("Got signal:", sig.String())
			return
		}
	}
}
