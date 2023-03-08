package main

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/clients"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"time"
)

func main() {
	// Creating objects of metrics and client
	mtrcs := metrics.NewMetrics()
	client := clients.NewMetricsClient()

	// Creating poll and report intervals
	pollInterval := time.NewTicker(2 * time.Second)
	reportInterval := time.NewTicker(10 * time.Second)

	// Agent's process
	for {
		select {
		case <-pollInterval.C:
			// Updating metrics
			metrics.UpdateMetrics(mtrcs)
		case <-reportInterval.C:
			// Sending metrics
			clients.MetricsUpload(client, mtrcs)
		}
	}
}
