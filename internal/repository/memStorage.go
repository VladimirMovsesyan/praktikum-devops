package repository

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"log"
)

type MemStorage struct {
	mtrcs map[string]metrics.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		mtrcs: map[string]metrics.Metric{},
	}
}

func (ms *MemStorage) GetMetrics() map[string]metrics.Metric {
	return ms.mtrcs
}

func (ms *MemStorage) Update(newMetric metrics.Metric) {
	switch newMetric.GetKind() {
	case "gauge":
		ms.mtrcs[newMetric.GetName()] = newMetric
	case "counter":
		metric, ok := ms.mtrcs[newMetric.GetName()]
		if ok {
			ms.mtrcs[newMetric.GetName()] = metrics.NewMetricCounter(
				newMetric.GetName(),
				metric.GetCounterValue()+newMetric.GetCounterValue(),
			)
		} else {
			ms.mtrcs[newMetric.GetName()] = newMetric
		}
	default:
		log.Println("Error: not implemented!")
	}
}
