package repository

import (
	"database/sql"
	"errors"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	_ "github.com/lib/pq"
	"log"
	"strconv"
	"strings"
	"time"
)

// PostgreStorage - contains pointer to pool of db connections.
type PostgreStorage struct {
	db *sql.DB
}

func NewPostgreStorage(db *sql.DB) *PostgreStorage {
	storage := &PostgreStorage{db: db}
	storage.ensureTableExists()
	return storage
}

type dbMetric struct {
	name  string
	mType string
	delta sql.NullInt64
	value sql.NullFloat64
}

var _ sql.Scanner = &dbMetric{}

func (metric *dbMetric) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		log.Println("couldn't convert value to []byte")
		return errors.New("couldn't convert value to []byte")
	}

	sv := string(bytes)
	split := strings.Split(sv[1:len(sv)-1], ",")

	metric.name = split[0]
	metric.mType = split[1]

	switch metric.mType {
	case "gauge":
		value, err := strconv.ParseFloat(split[3], 64)
		if err != nil {
			log.Println(err)
			return err
		}
		metric.value.Float64, metric.value.Valid = value, true
	case "counter":
		delta, err := strconv.ParseInt(split[2], 10, 64)
		if err != nil {
			log.Println(err)
			return err
		}
		metric.delta.Int64, metric.delta.Valid = delta, true
	default:
		log.Println("not implemented")
		return errors.New("not implemented")
	}

	return nil
}

func (storage *PostgreStorage) GetMetricsMap() map[string]metrics.Metric {
	metricsMap := make(map[string]metrics.Metric)

	rows, err := storage.db.Query(`SELECT (metric_name, metric_type, metric_delta, metric_value) FROM metric`)
	if err != nil || rows.Err() != nil {
		return nil
	}

	for rows.Next() {
		var dbObj dbMetric
		err = rows.Scan(&dbObj)
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
	row := storage.db.QueryRow(
		`SELECT (metric_name, metric_type, metric_delta, metric_value) FROM metric WHERE metric_name = $1`,
		name,
	)
	if row.Err() != nil {
		return metrics.Metric{}, row.Err()
	}

	var dbObj dbMetric
	err := row.Scan(&dbObj)
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

func (storage *PostgreStorage) BatchUpdate(metrics []metrics.Metric) {
	tx, err := storage.db.Begin()
	if err != nil {
		log.Println(err)
		return
	}
	defer tx.Rollback()

	selectStmt, err := tx.Prepare(`SELECT COUNT(*) FROM metric WHERE metric_name = $1`)
	if err != nil {
		log.Println(err)
		return
	}

	insertGaugeStmt, err := tx.Prepare(`INSERT INTO metric (metric_name, metric_type, metric_value, created_at, updated_at) 
												VALUES ($1, 'gauge', $2, $3, $4)`)
	if err != nil {
		log.Println(err)
		return
	}

	insertCounterStmt, err := tx.Prepare(`INSERT INTO metric (metric_name, metric_type, metric_delta, created_at, updated_at) 
									VALUES ($1, 'counter', $2, $3, $4)`)
	if err != nil {
		log.Println(err)
		return
	}

	updateGaugeStmt, err := tx.Prepare(
		`UPDATE metric SET metric_value=$1, updated_at=$2 WHERE metric_name=$3`,
	)
	if err != nil {
		log.Println(err)
		return
	}

	updateCounterStmt, err := tx.Prepare(
		`UPDATE metric SET metric_delta=metric_delta+$1, updated_at=$2 WHERE metric_name=$3`,
	)
	if err != nil {
		log.Println(err)
		return
	}

	for _, metric := range metrics {
		row := selectStmt.QueryRow(metric.GetName())
		if row.Err() != nil {
			log.Println(err)
			return
		}

		var count int
		err = row.Scan(&count)
		if err != nil {
			log.Println(err)
			return
		}

		switch metric.GetKind() {
		case "gauge":
			if count == 0 {
				_, err = insertGaugeStmt.Exec(metric.GetName(), metric.GetGaugeValue(), time.Now(), time.Now())
				if err != nil {
					log.Println(err)
					return
				}
			} else {
				_, err = updateGaugeStmt.Exec(metric.GetGaugeValue(), time.Now(), metric.GetName())
				if err != nil {
					log.Println(err)
					return
				}
			}
		case "counter":
			if count == 0 {
				_, err = insertCounterStmt.Exec(metric.GetName(), metric.GetCounterValue(), time.Now(), time.Now())
				if err != nil {
					log.Println(err)
					return
				}
			} else {
				_, err = updateCounterStmt.Exec(metric.GetCounterValue(), time.Now(), metric.GetName())
				if err != nil {
					log.Println(err)
					return
				}
			}
		default:
			log.Fatal("not implemented")
		}
	}

	tx.Commit()
}
