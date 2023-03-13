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

func MetricsUpload(client *http.Client, mtrcs *metrics.Metrics, baseURL string) {
	for _, metric := range mtrcs.MetricSlice {
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
		}

		err = resp.Body.Close()
		if err != nil {
			log.Println("Couldn't close body of response. Error: ", err)
		}
	}
	mtrcs.ResetPollCounter()
}
