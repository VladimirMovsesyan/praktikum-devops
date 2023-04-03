package main

import (
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
	defaultPoll   = 2
	defaultReport = 10
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Creating objects of metrics and client
	mtrcs := metrics.NewMetrics()

	// Creating poll and report intervals
	pollInterval := time.NewTicker(
		utils.UpdateInterval(
			"POLL_INTERVAL",
			defaultPoll,
		) * time.Second,
	)
	reportInterval := time.NewTicker(
		utils.UpdateInterval(
			"REPORT_INTERVAL",
			defaultReport,
		) * time.Second,
	)

	// Agent's process
	for {
		select {
		case <-pollInterval.C:
			//Updating metrics
			go metrics.UpdateMetrics(mtrcs)
		case <-reportInterval.C:
			//Sending metrics
			go clients.MetricsUpload(mtrcs)
		case sig := <-signals:
			log.Println(sig.String())
			return
		}
	}
}
