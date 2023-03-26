package clients

import (
	"fmt"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/utils"
	"log"
	"net/http"
	"time"
)

const (
	defaultProtocol     = "http://"
	updateGaugeFormat   = "/update/%s/%s/%f"
	updateCounterFormat = "/update/%s/%s/%d"
)

func NewMetricsClient() *http.Client {
	client := &http.Client{}
	client.Timeout = 3 * time.Second
	return client
}

func MetricsUpload(mtrcs *metrics.Metrics) {
	address := defaultProtocol + utils.UpdateAddress("ADDRESS", utils.DefaultAddress)
	for _, metric := range mtrcs.MetricSlice {
		metricUpload(address, metric)
	}
	mtrcs.ResetPollCounter()
}

func metricUpload(address string, metric metrics.Metric) {
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

	resp, err := client.Post(url, "text/plain", nil)
	if err != nil {
		log.Println("Error: ", err)
		return
	}
	defer resp.Body.Close()
}
