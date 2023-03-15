package clients

import (
	"fmt"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"log"
	"net/http"
	"time"
)

const (
	updateGaugeFormat   = "/update/%s/%s/%f"
	updateCounterFormat = "/update/%s/%s/%d"
)

func NewMetricsClient() *http.Client {
	client := &http.Client{}
	client.Timeout = 3 * time.Second
	return client
}

func MetricsUpload(mtrcs *metrics.Metrics) {
	for _, metric := range mtrcs.MetricSlice {
		metricUpload("http://127.0.0.1:8080", metric)
	}
	mtrcs.ResetPollCounter()
}

func metricUpload(baseURL string, metric metrics.Metric) {
	client := NewMetricsClient()
	url := baseURL
	switch metric.GetKind() {
	case "gauge":
		url = fmt.Sprintf(url+updateGaugeFormat, metric.GetKind(), metric.GetName(), metric.GetGaugeValue())
	case "counter":
		url = fmt.Sprintf(url+updateCounterFormat, metric.GetKind(), metric.GetName(), metric.GetCounterValue())
	default:
		log.Fatal("Error: unsupported metric type!")
	}

	resp, err := client.Post(url, "text/plain", nil)
	if err != nil {
		log.Println("Error: ", err)
		return
	}

	defer func(resp *http.Response) {
		err := resp.Body.Close()
		if err != nil {
			log.Println("Couldn't close body of response. Error: ", err)
		}
	}(resp)
}
