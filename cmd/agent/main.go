package main

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/clients"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		// Creating objects of metrics and client
		mtrcs := metrics.NewMetrics()

		// Creating poll and report intervals
		pollInterval := time.NewTicker(2 * time.Second)
		reportInterval := time.NewTicker(10 * time.Second)

		// Agent's process
		for {
			select {
			case <-pollInterval.C:
				//Updating metrics
				wg.Add(1)
				go metrics.UpdateMetrics(mtrcs)
				wg.Done()
			case <-reportInterval.C:
				//Sending metrics
				wg.Add(1)
				go clients.MetricsUpload(mtrcs)
				wg.Done()
			case sig := <-signals:
				log.Println(sig.String())
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
}
