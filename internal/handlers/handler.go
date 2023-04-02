package handlers

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type MetricRepository interface {
	GetMetricsMap() map[string]metrics.Metric
	GetMetric(name string) (metrics.Metric, error)
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

type JSONMetric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewJSONMetric(metric metrics.Metric) (*JSONMetric, error) {
	jsonMetric := &JSONMetric{
		ID:    metric.GetName(),
		MType: metric.GetKind(),
	}

	switch metric.GetKind() {
	case "gauge":
		value := float64(metric.GetGaugeValue())
		jsonMetric.Value = &value
	case "counter":
		delta := int64(metric.GetCounterValue())
		jsonMetric.Delta = &delta
	default:
		return nil, errors.New("not implemented type")
	}

	return jsonMetric, nil
}

func JSONUpdateHandler(storage MetricRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer r.Body.Close()
		var jsonMetric JSONMetric

		err = json.Unmarshal(bytes, &jsonMetric)
		if err != nil {
			log.Println(err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch jsonMetric.MType {
		case "gauge":
			storage.Update(
				metrics.NewMetricGauge(
					jsonMetric.ID,
					metrics.Gauge(*jsonMetric.Value),
				),
			)
		case "counter":
			storage.Update(
				metrics.NewMetricCounter(
					jsonMetric.ID,
					metrics.Counter(*jsonMetric.Delta),
				),
			)
		default:
			log.Fatal("Not implemented")
		}

		rw.Header().Add("Content-Type", "application/json")

		_, err = rw.Write(bytes)
		if err != nil {
			log.Println(err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
	}
}

func updateJSONMetric(jsonMetric *JSONMetric, metric metrics.Metric) {
	switch metric.GetKind() {
	case "gauge":
		value := float64(metric.GetGaugeValue())
		jsonMetric.Value = &value
	case "counter":
		delta := int64(metric.GetCounterValue())
		jsonMetric.Delta = &delta
	default:
		log.Fatal("Not implemented")
	}
}

func JSONPrintHandler(storage MetricRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer r.Body.Close()
		var jsonMetric JSONMetric

		err = json.Unmarshal(bytes, &jsonMetric)
		if err != nil {
			log.Println(err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		metric, err := storage.GetMetric(jsonMetric.ID)
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		updateJSONMetric(&jsonMetric, metric)

		marshal, err := json.Marshal(jsonMetric)
		if err != nil {
			log.Println(err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.Header().Add("Content-Type", "application/json")
		_, err = rw.Write(marshal)
		if err != nil {
			log.Println(err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
	}
}

func PrintStorageHandler(storage MetricRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mtrcs := storage.GetMetricsMap()
		for _, value := range mtrcs {
			result := value.GetKind() + " " + value.GetName() + " "
			switch value.GetKind() {
			case "gauge":
				result += fmt.Sprintf("%.3f", value.GetGaugeValue())
			case "counter":
				result += fmt.Sprintf("%d", value.GetCounterValue())
			}
			rw.Header().Set("Content-Type", "text/html")
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

		value, err := storage.GetMetric(name)
		if err != nil || value.GetKind() != kind {
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

		_, err = rw.Write([]byte(result))
		if err != nil {
			log.Println("Error: Couldn't write data to response!")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
	}
}

type gzipWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (gw gzipWriter) Write(p []byte) (int, error) {
	return gw.writer.Write(p)
}

func CompressHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(rw, r)
				return
			}

			newWriter := gzip.NewWriter(rw)
			defer newWriter.Close()

			writer := gzipWriter{
				ResponseWriter: rw,
				writer:         newWriter,
			}

			writer.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(writer, r)
		},
	)
}

func DecompressHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				next.ServeHTTP(rw, r)
				return
			}

			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Println(err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer reader.Close()

			r.Body = reader
			next.ServeHTTP(rw, r)
		},
	)
}
