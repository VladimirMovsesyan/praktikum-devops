package repository

import (
	"errors"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"log"
	"sync"
)

type MemStorage struct {
	mutex sync.RWMutex
	mtrcs map[string]metrics.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		mtrcs: map[string]metrics.Metric{},
	}
}

func (ms *MemStorage) GetMetricsMap() map[string]metrics.Metric {
	return ms.mtrcs
}

func (ms *MemStorage) GetMetric(name string) (metrics.Metric, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	metric, ok := ms.mtrcs[name]
	if !ok {
		return metrics.Metric{}, errors.New("metric not found")
	}

	return metric, nil
}

func (ms *MemStorage) Update(newMetric metrics.Metric) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

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

func (ms *MemStorage) BatchUpdate(metrics []metrics.Metric) {
	for _, metric := range metrics {
		ms.Update(metric)
	}
}
