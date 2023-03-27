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
		time.Duration(
			utils.UpdateIntVar(
				"POLL_INTERVAL",
				"p",
				defaultPoll,
				"Interval of polling metrics",
			),
		) * time.Second,
	)
	reportInterval := time.NewTicker(
		time.Duration(
			utils.UpdateIntVar(
				"REPORT_INTERVAL",
				"r",
				defaultReport,
				"Interval of reporting metrics to server",
			),
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
