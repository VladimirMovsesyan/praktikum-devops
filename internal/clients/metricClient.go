package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/crypt"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/handlers"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/hash"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
	"log"
	"net/http"
	"time"
)

const (
	defaultProtocol     = "http://"
	updateGaugeFormat   = "/update/%s/%s/%f"
	hashGaugeFormat     = "%s:%s:%f"
	updateCounterFormat = "/update/%s/%s/%d"
	hashCounterFormat   = "%s:%s:%d"
)

func NewMetricsClient() *http.Client {
	client := &http.Client{}
	client.Timeout = 3 * time.Second
	return client
}

func MetricsUpload(storage metricRepository, address, key, cryptoPath string) {
	address = defaultProtocol + address

	log.Println("sending metrics to:", address)
	metricsUpload(storage, address, key, cryptoPath)

	metrics.ResetPollCounter(storage)
}

func metricsUpload(storage metricRepository, address, key, cryptoPath string) {
	client := NewMetricsClient()
	url := address + "/updates/"

	metricsMap := storage.GetMetricsMap()

	if len(metricsMap) == 0 {
		log.Println("Empty batch, uploading skipped")
		return
	}

	jsonMetrics := make([]handlers.JSONMetric, 0, len(metricsMap))

	for _, metric := range metricsMap {
		var jsonMetric handlers.JSONMetric

		switch metric.GetKind() {
		case "gauge":
			value := float64(metric.GetGaugeValue())
			jsonMetric = handlers.JSONMetric{
				ID:    metric.GetName(),
				MType: metric.GetKind(),
				Value: &value,
			}
		case "counter":
			delta := int64(metric.GetCounterValue())
			jsonMetric = handlers.JSONMetric{
				ID:    metric.GetName(),
				MType: metric.GetKind(),
				Delta: &delta,
			}
		default:
			log.Fatal("not implemented")
		}

		hashData, err := getHashData(metric)
		if err != nil {
			return
		}

		if key != "" {
			jsonMetric.Hash = hash.Get(hashData, key)
		}
		jsonMetrics = append(jsonMetrics, jsonMetric)
	}

	marshal, err := json.Marshal(&jsonMetrics)
	if err != nil {
		log.Println("Error: ", err)
		return
	}

	if cryptoPath != "" {
		c, err := crypt.New(crypt.WithPublicKey(cryptoPath))
		if err != nil {
			log.Println(err)
		}

		marshal, err = c.Encrypt(marshal)
		if err != nil {
			log.Println(err)
		}
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(marshal))
	if err != nil {
		log.Println("Error: ", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error: ", err)
		return
	}
	defer resp.Body.Close()
}

func metricUpload(address string, metric metrics.Metric, key string) {
	client := NewMetricsClient()

	url := address

	switch metric.GetKind() {
	case "gauge":
		url = fmt.Sprintf(url+updateGaugeFormat, metric.GetKind(), metric.GetName(), metric.GetGaugeValue())
	case "counter":
		url = fmt.Sprintf(url+updateCounterFormat, metric.GetKind(), metric.GetName(), metric.GetCounterValue())
	default:
		log.Fatal("Error: unsupported metric type!")
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Println("Error: ", err)
		return
	}

	hashData, err := getHashData(metric)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "text/plain")
	if key != "" {
		req.Header.Set("Hash", hash.Get(hashData, key))
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error: ", err)
		return
	}
	defer resp.Body.Close()
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

type metricRepository interface {
	GetMetricsMap() map[string]metrics.Metric
	GetMetric(name string) (metrics.Metric, error)
	Update(metrics.Metric)
	BatchUpdate(metrics []metrics.Metric)
}

type workerPool struct {
	workerCnt  int
	address    string
	key        string
	cryptoPath string
	storage    metricRepository
	taskCh     chan string
}

func NewWorkerPool(workerCnt int, address, key, cryptoPath string) *workerPool {
	return &workerPool{
		workerCnt:  workerCnt,
		address:    address,
		key:        key,
		cryptoPath: cryptoPath,
		storage:    repository.NewMemStorage(),
		taskCh:     make(chan string),
	}
}

func (wp *workerPool) Run() {
	for i := 0; i < wp.workerCnt; i++ {
		go func() {
			for task := range wp.taskCh {
				switch task {
				case "updateMem":
					metrics.UpdateMetrics(wp.storage)
				case "updateGopsutil":
					metrics.UpdateMetricsGopsutil(wp.storage)
				case "upload":
					MetricsUpload(wp.storage, wp.address, wp.key, wp.cryptoPath)
				default:
					log.Println("not implemented type of worker pool's task")
				}
			}
		}()
	}
}

func (wp *workerPool) AddTask(task string) {
	wp.taskCh <- task
}

func (wp *workerPool) Stop() {
	close(wp.taskCh)
}
