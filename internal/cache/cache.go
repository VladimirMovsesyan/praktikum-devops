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

type importer struct {
	file    *os.File
	decoder *json.Decoder
}

func newImporter(filename string) (*importer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &importer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (imp *importer) importStorage(storage metricRepository) error {
	for {
		var jsonMetric handlers.JSONMetric
		err := imp.decoder.Decode(&jsonMetric)
		if errors.Is(err, io.EOF) {
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
	imp, err := newImporter(filename)
	if err != nil {
		return err
	}
	defer imp.close()

	err = imp.importStorage(storage)
	return err
}

func (imp *importer) close() error {
	return imp.file.Close()
}

type exporter struct {
	file    *os.File
	encoder *json.Encoder
}

func newExporter(filename string) (*exporter, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &exporter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (exp *exporter) exportStorage(storage metricRepository) error {
	metricMap := storage.GetMetricsMap()

	for _, value := range metricMap {
		err := exp.exportEvent(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (exp *exporter) exportEvent(metric metrics.Metric) error {
	jsonMetric, err := handlers.NewJSONMetric(metric)
	if err != nil {
		return err
	}

	err = exp.encoder.Encode(&jsonMetric)
	return err
}

func (exp *exporter) close() error {
	return exp.file.Close()
}

func ExportData(filename string, storage metricRepository) error {
	exp, err := newExporter(filename)
	if err != nil {
		return err
	}
	defer exp.close()

	err = exp.exportStorage(storage)
	return err
}
