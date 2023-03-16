package repository

import "github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"

type MetricRepository interface {
	GetMetrics() []metrics.Metric
	Update(metrics.Metric)
	Delete(metrics.Metric)
}
