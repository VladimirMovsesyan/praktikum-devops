package repository

import (
	"database/sql"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type PostgreStorage struct {
	db *sql.DB
}

func NewPostgreStorage(dbDSN string) *PostgreStorage {
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		return nil
	}
	return &PostgreStorage{db: db}
}

type dbMetric struct {
	name  string
	mType string
	delta sql.NullInt64
	value sql.NullFloat64
}

func (storage *PostgreStorage) GetMetricsMap() map[string]metrics.Metric {
	storage.ensureTableExists()
	metricsMap := make(map[string]metrics.Metric)

	rows, err := storage.db.Query(`SELECT metric_name, metric_type, metric_delta, metric_value FROM metric`)
	if err != nil || rows.Err() != nil {
		return nil
	}

	for rows.Next() {
		var dbObj dbMetric
		err = rows.Scan(&dbObj.name, &dbObj.mType, &dbObj.delta, &dbObj.value)
		if err != nil {
			return nil
		}

		var metric metrics.Metric
		switch dbObj.mType {
		case "gauge":
			metric = metrics.NewMetricGauge(dbObj.name, metrics.Gauge(dbObj.value.Float64))
		case "counter":
			metric = metrics.NewMetricCounter(dbObj.name, metrics.Counter(dbObj.delta.Int64))
		default:
			log.Fatal("not implemented")
		}

		metricsMap[metric.GetName()] = metric
	}

	return metricsMap
}

func (storage *PostgreStorage) GetMetric(name string) (metrics.Metric, error) {
	storage.ensureTableExists()
	row := storage.db.QueryRow(
		`SELECT metric_name, metric_type, metric_delta, metric_value FROM metric WHERE metric_name = $1`,
		name,
	)
	if row.Err() != nil {
		return metrics.Metric{}, row.Err()
	}

	var dbObj dbMetric
	err := row.Scan(&dbObj.name, &dbObj.mType, &dbObj.delta, &dbObj.value)
	if err != nil {
		return metrics.Metric{}, err
	}

	var metric metrics.Metric
	switch dbObj.mType {
	case "gauge":
		metric = metrics.NewMetricGauge(dbObj.name, metrics.Gauge(dbObj.value.Float64))
	case "counter":
		metric = metrics.NewMetricCounter(dbObj.name, metrics.Counter(dbObj.delta.Int64))
	default:
		log.Fatal("not implemented")
	}

	return metric, nil
}

func (storage *PostgreStorage) Update(metric metrics.Metric) {
	storage.ensureTableExists()
	row := storage.db.QueryRow(`SELECT COUNT(*) FROM metric WHERE metric_name = $1`, metric.GetName())
	if row.Err() != nil {
		return
	}

	var count int
	err := row.Scan(&count)
	if err != nil {
		return
	}

	if count == 0 {
		storage.insertMetric(metric)
		return
	}
	storage.updateMetric(metric)
}

func (storage *PostgreStorage) insertMetric(metric metrics.Metric) {
	switch metric.GetKind() {
	case "gauge":
		_, _ = storage.db.Exec(
			`INSERT INTO metric (metric_name, metric_type, metric_value, created_at, updated_at) 
					VALUES ($1, 'gauge', $2, $3, $4)`,
			metric.GetName(),
			metric.GetGaugeValue(),
			time.Now(),
			time.Now(),
		)
	case "counter":
		_, _ = storage.db.Exec(
			`INSERT INTO metric (metric_name, metric_type, metric_delta, created_at, updated_at) 
					VALUES ($1, 'counter', $2, $3, $4)`,
			metric.GetName(),
			metric.GetCounterValue(),
			time.Now(),
			time.Now(),
		)
	default:
		log.Fatal("not implemented")
	}
}

func (storage *PostgreStorage) updateMetric(metric metrics.Metric) {
	switch metric.GetKind() {
	case "gauge":
		_, _ = storage.db.Exec(
			`UPDATE metric SET metric_value=$1, updated_at=$2 WHERE metric_name=$3`,
			metric.GetGaugeValue(),
			time.Now(),
			metric.GetName(),
		)
	case "counter":
		_, _ = storage.db.Exec(
			`UPDATE metric SET metric_delta=metric_delta+$1, updated_at=$2 WHERE metric_name=$3`,
			metric.GetCounterValue(),
			time.Now(),
			metric.GetName(),
		)
	default:
		log.Fatal("not implemented")
	}
}

const tableCreation string = `CREATE TABLE IF NOT EXISTS metric (
    metric_name VARCHAR (50) PRIMARY KEY, 
    metric_type VARCHAR (10) NOT NULL, 
    metric_delta BIGINT, 
    metric_value DOUBLE PRECISION, 
    created_at TIMESTAMP NOT NULL, 
    updated_at TIMESTAMP NOT NULL
);`

func (storage *PostgreStorage) ensureTableExists() {
	_, _ = storage.db.Exec(tableCreation)
}
