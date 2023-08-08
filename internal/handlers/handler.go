package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/hash"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/proto"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"strconv"
)

type metricRepository interface {
	GetMetricsMap() map[string]metrics.Metric
	GetMetric(name string) (metrics.Metric, error)
	Update(metrics.Metric)
	BatchUpdate(metrics []metrics.Metric)
}

const (
	hashGaugeFormat   = "%s:%s:%f"
	hashCounterFormat = "%s:%s:%d"
)

// UpdateStorageHandler - handler that routing from "/update/kind/name/value".
// Parsing query params to values and updating metric in DB.
// If metric with such name and kind doesn't exist, creating new metric.
func UpdateStorageHandler(storage metricRepository, key string) http.HandlerFunc {
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

		hashData, err := getHashData(newMetric)
		if err != nil {
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}

		hashHeader := r.Header.Get("Hash")
		if key != "" {
			if !hash.Valid(hashHeader, hashData, key) {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			rw.Header().Set("Hash", hash.Get(hashData, key))
		}

		storage.Update(newMetric)
		rw.WriteHeader(http.StatusOK)
	}
}

// JSONMetric - struct that helps to marshal/unmarshal metric to/from json representation.
type JSONMetric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
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

// JSONUpdateHandler - handler that routing from "/update".
// Parsing json provided data to values and updating metric in DB.
// If metric with such name and kind doesn't exist, creating new metric.
func JSONUpdateHandler(storage metricRepository, key string) http.HandlerFunc {
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

		var metric metrics.Metric

		switch jsonMetric.MType {
		case "gauge":
			metric = metrics.NewMetricGauge(
				jsonMetric.ID,
				metrics.Gauge(*jsonMetric.Value),
			)
		case "counter":
			metric = metrics.NewMetricCounter(
				jsonMetric.ID,
				metrics.Counter(*jsonMetric.Delta),
			)
		default:
			log.Fatal("Not implemented")
		}

		hashData, err := getHashData(metric)
		if err != nil {
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}

		hashHeader := jsonMetric.Hash
		if key != "" {
			if !hash.Valid(hashHeader, hashData, key) {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			rw.Header().Set("Hash", hash.Get(hashData, key))
		}

		storage.Update(metric)

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

// MetricsUpdateHandler - handler that routing from "/updates".
// Parsing json provided batch of metric to values and updating metrics in DB.
// If metrics with such name and kind doesn't exist, creating new metric.
func MetricsUpdateHandler(storage metricRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var jsonSlice []JSONMetric

		err = json.Unmarshal(bytes, &jsonSlice)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		metricSlice := make([]metrics.Metric, 0, len(jsonSlice))

		for _, jsonMetric := range jsonSlice {
			var metric metrics.Metric

			switch jsonMetric.MType {
			case "gauge":
				metric = metrics.NewMetricGauge(
					jsonMetric.ID,
					metrics.Gauge(*jsonMetric.Value),
				)
			case "counter":
				metric = metrics.NewMetricCounter(
					jsonMetric.ID,
					metrics.Counter(*jsonMetric.Delta),
				)
			default:
				log.Fatal("not implemented")
			}

			metricSlice = append(metricSlice, metric)
		}

		storage.BatchUpdate(metricSlice)

		rw.Header().Set("Content-Type", "application/json")

		_, err = rw.Write(bytes)
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
	}
}

func GRPCMetricUpdateHandler(storage metricRepository) func(ctx context.Context, request *proto.BatchUpdateMetricsRequest) (*proto.BatchUpdateMetricsResponse, error) {
	return func(ctx context.Context, in *proto.BatchUpdateMetricsRequest) (*proto.BatchUpdateMetricsResponse, error) {
		for _, m := range in.GetMetrics() {
			var metric metrics.Metric
			switch m.MType.String() {
			case "counter":
				metric = metrics.NewMetricCounter(m.ID, metrics.Counter(m.GetDelta()))
			case "gauge":
				metric = metrics.NewMetricGauge(m.ID, metrics.Gauge(m.GetValue()))
			default:
				return nil, errors.New("unsupported metric type")
			}
			storage.Update(metric)
		}
		response := &proto.BatchUpdateMetricsResponse{}
		return response, nil
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

// JSONPrintHandler - handler that routing from "/value".
// Parsing json provided data that contains metric name and kind and returning metric with such name and kind.
// If metric with such name and kind doesn't exist returning status 404.
func JSONPrintHandler(storage metricRepository, key string) http.HandlerFunc {
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

		hashData, err := getHashData(metric)
		if err != nil {
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}

		jsonMetric.Hash = hash.Get(hashData, key)
		hashHeader := jsonMetric.Hash
		if key != "" {
			if !hash.Valid(hashHeader, hashData, key) {
				log.Println(hashData)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			rw.Header().Set("Hash", hash.Get(hashData, key))
		}

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

// PrintStorageHandler - handler that routing from "/".
// Printing all metrics from db.
func PrintStorageHandler(storage metricRepository) http.HandlerFunc {
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

// PrintValueHandler - handler that routing from "/value/kind/name".
// Parsing metrics kind and name from query params and printing metric with such kind and name.
// If metric with such name and kind doesn't exist, returning status 404.
func PrintValueHandler(storage metricRepository, key string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		kind := chi.URLParam(r, "kind")
		name := chi.URLParam(r, "name")

		metric, err := storage.GetMetric(name)
		if err != nil || metric.GetKind() != kind {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		var result string

		switch kind {
		case "gauge":
			result = fmt.Sprintf("%.3f", metric.GetGaugeValue())
		case "counter":
			result = fmt.Sprintf("%d", metric.GetCounterValue())
		default:
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}

		hashData, err := getHashData(metric)
		if err != nil {
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}

		hashHeader := r.Header.Get("Hash")
		if key != "" {
			if !hash.Valid(hashHeader, hashData, key) {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			rw.Header().Set("Hash", hash.Get(hashData, key))
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

func getHashData(metric metrics.Metric) (string, error) {
	var hashData string
	switch metric.GetKind() {
	case "gauge":
		hashData = fmt.Sprintf(hashGaugeFormat, metric.GetName(), metric.GetKind(), metric.GetGaugeValue())
	case "counter":
		hashData = fmt.Sprintf(hashCounterFormat, metric.GetName(), metric.GetKind(), metric.GetCounterValue())
	default:
		log.Println("not implemented type")
		return "", errors.New("not implemented type")
	}
	return hashData, nil
}

// PingDatabaseHandler - handler that routing from "/ping".
// Testing connection to DB.
func PingDatabaseHandler(db *sql.DB) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		err := db.Ping()
		if err != nil {
			log.Println("Couldn't ping database")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
	}
}
