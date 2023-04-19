package cache

import (
	"encoding/json"
	"errors"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/handlers"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"io"
	"os"
)

type metricRepository interface {
	GetMetricsMap() map[string]metrics.Metric
	Update(metrics.Metric)
}

type Importer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewImporter(filename string) (*Importer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &Importer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (imp *Importer) Import(storage metricRepository) error {
	for {
		var jsonMetric handlers.JSONMetric
		err := imp.decoder.Decode(&jsonMetric)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		var metric metrics.Metric

		switch jsonMetric.MType {
		case "gauge":
			metric = metrics.NewMetricGauge(jsonMetric.ID, metrics.Gauge(*jsonMetric.Value))
		case "counter":
			metric = metrics.NewMetricCounter(jsonMetric.ID, metrics.Counter(*jsonMetric.Delta))
		default:
			return errors.New("not implemented type")
		}

		storage.Update(metric)
	}

	return nil
}

func ImportData(filename string, storage metricRepository) error {
	importer, err := NewImporter(filename)
	if err != nil {
		return err
	}
	defer importer.Close()

	err = importer.Import(storage)
	return err
}

func (imp *Importer) Close() error {
	return imp.file.Close()
}

type Exporter struct {
	file    *os.File
	encoder *json.Encoder
}

func NewExporter(filename string) (*Exporter, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (exp *Exporter) ExportStorage(storage metricRepository) error {
	metricMap := storage.GetMetricsMap()

	for _, value := range metricMap {
		err := exp.exportEvent(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (exp *Exporter) exportEvent(metric metrics.Metric) error {
	jsonMetric, err := handlers.NewJSONMetric(metric)
	if err != nil {
		return err
	}

	err = exp.encoder.Encode(&jsonMetric)
	return err
}

func (exp *Exporter) Close() error {
	return exp.file.Close()
}

func ExportData(filename string, storage metricRepository) error {
	exporter, err := NewExporter(filename)
	if err != nil {
		return err
	}
	defer exporter.Close()

	err = exporter.ExportStorage(storage)
	return err
}
