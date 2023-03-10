package main

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func updateStorageHandler(storage repository.MetricRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		urlSlice := strings.Split(r.URL.Path, "/")
		if len(urlSlice) != 5 {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		kind, name := urlSlice[2], urlSlice[3]
		newMetric := metrics.Metric{}

		switch kind {
		case "gauge":
			value, err := strconv.Atoi(urlSlice[4])
			if err != nil {
				log.Println(err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			newMetric = metrics.NewMetricGauge(name, metrics.Gauge(value))
		case "counter":
			value, err := strconv.Atoi(urlSlice[4])
			if err != nil {
				log.Println(err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			newMetric = metrics.NewMetricCounter(name, metrics.Counter(value))
		default:
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}
		storage.Update(newMetric)
		rw.WriteHeader(http.StatusOK)
	}
}

func main() {
	storage := &repository.MemStorage{}
	http.HandleFunc("/update/", updateStorageHandler(storage))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
