package clients

import (
	"fmt"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/hash"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/utils"
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

func MetricsUpload(mtrcs *metrics.Metrics, flAddr *string, key string) {
	address := defaultProtocol + utils.UpdateStringVar(
		"ADDRESS",
		flAddr,
	)

	log.Println("sending metrics to:", address)
	for _, metric := range mtrcs.MetricSlice {
		metricUpload(address, metric, key)
	}

	mtrcs.ResetPollCounter()
}

func metricUpload(address string, metric metrics.Metric, key string) {
	client := NewMetricsClient()

	url := address
	var hashData string

	switch metric.GetKind() {
	case "gauge":
		url = fmt.Sprintf(url+updateGaugeFormat, metric.GetKind(), metric.GetName(), metric.GetGaugeValue())
		hashData = fmt.Sprintf(hashGaugeFormat, metric.GetName(), metric.GetKind(), metric.GetGaugeValue())
	case "counter":
		url = fmt.Sprintf(url+updateCounterFormat, metric.GetKind(), metric.GetName(), metric.GetCounterValue())
		hashData = fmt.Sprintf(hashCounterFormat, metric.GetName(), metric.GetKind(), metric.GetCounterValue())
	default:
		log.Fatal("Error: unsupported metric type!")
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Println("Error: ", err)
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
