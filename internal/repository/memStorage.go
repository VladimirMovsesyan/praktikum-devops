package repository

import "github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"

type MemStorage struct {
	mtrcs []metrics.Metric
}

func sameMetric(metric1, metric2 metrics.Metric) bool {
	return metric1.GetKind() == metric2.GetKind() && metric1.GetName() == metric2.GetName()
}

func (ms *MemStorage) GetMetrics() []metrics.Metric {
	return ms.mtrcs
}

func (ms *MemStorage) Update(newMetric metrics.Metric) {
	for index, metric := range ms.mtrcs {
		if sameMetric(metric, newMetric) {
			switch metric.GetKind() {
			case "gauge":
				ms.mtrcs[index] = newMetric
			case "counter":
				ms.mtrcs[index] = metrics.NewMetricCounter(
					metric.GetName(),
					metric.GetCounterValue()+newMetric.GetCounterValue(),
				)
			}
			return
		}
	}
	ms.mtrcs = append(ms.mtrcs, newMetric)
}

func (ms *MemStorage) Delete(newMetric metrics.Metric) {
	for i, metric := range ms.mtrcs {
		if sameMetric(metric, newMetric) {
			ms.mtrcs = append(ms.mtrcs[:i], ms.mtrcs[i+1:]...)
			return
		}
	}
}
