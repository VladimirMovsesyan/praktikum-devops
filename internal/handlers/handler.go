package handlers

import (
	"fmt"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

type MetricRepository interface {
	GetMetrics() map[string]metrics.Metric
	Update(metrics.Metric)
}

func UpdateStorageHandler(storage MetricRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		kind := chi.URLParam(r, "kind")
		name := chi.URLParam(r, "name")
		value := chi.URLParam(r, "value")

		var newMetric metrics.Metric

		if value == "" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		switch kind {
		case "gauge":
			value, err := strconv.ParseFloat(value, 64)
			if err != nil {
				log.Println(err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			newMetric = metrics.NewMetricGauge(name, metrics.Gauge(value))
		case "counter":
			value, err := strconv.Atoi(value)
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

func PrintStorageHandler(storage MetricRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mtrcs := storage.GetMetrics()
		for _, value := range mtrcs {
			result := value.GetKind() + " " + value.GetName() + " "
			switch value.GetKind() {
			case "gauge":
				result += fmt.Sprintf("%.3f", value.GetGaugeValue())
			case "counter":
				result += fmt.Sprintf("%d", value.GetCounterValue())
			}
			_, err := rw.Write([]byte(result))
			if err != nil {
				log.Println("Error: Couldn't write data to response!")
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		rw.WriteHeader(http.StatusOK)
	}
}

func PrintValueHandler(storage MetricRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		kind := chi.URLParam(r, "kind")
		name := chi.URLParam(r, "name")
		mtrcs := storage.GetMetrics()
		value, ok := mtrcs[name]

		if ok == false || value.GetKind() != kind {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		var result string

		switch kind {
		case "gauge":
			result = fmt.Sprintf("%.3f", value.GetGaugeValue())
		case "counter":
			result = fmt.Sprintf("%d", value.GetCounterValue())
		default:
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}

		_, err := rw.Write([]byte(result))
		if err != nil {
			log.Println("Error: Couldn't write data to response!")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
		return

	}
}
