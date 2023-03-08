package clients

import (
	"fmt"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"log"
	"net/http"
	"time"
)

const (
	scheme       = "http://"
	host         = "127.0.0.1:8080"
	updateFormat = "/update/%s/%s/%d"
)

func NewMetricsClient() *http.Client {
	client := &http.Client{}
	client.Timeout = 3 * time.Second
	return client
}

func MetricsUpload(client *http.Client, mtrcs *metrics.Metrics) {
	for _, metric := range mtrcs.MetricSlice {
		url := scheme + host + updateFormat
		switch metric.GetKind() {
		case "gauge":
			url = fmt.Sprintf(url, metric.GetKind(), metric.GetName(), int64(metric.GetGaugeValue()))
		case "counter":
			url = fmt.Sprintf(url, metric.GetKind(), metric.GetName(), int64(metric.GetCounterValue()))
		default:
			log.Println("Error: unsupported metric type!")
			url = fmt.Sprintf(url, metric.GetKind(), metric.GetName(), int64(metric.GetGaugeValue()))
		}

		_, err := client.Post(url, "text/plain", nil)
		if err != nil {
			log.Println("Error: ", err)
		}

	}
}
