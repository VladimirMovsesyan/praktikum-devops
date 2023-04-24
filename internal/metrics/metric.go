package metrics

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"
)

type Gauge float64
type Counter int64

type MetricKind int

const (
	GaugeKind MetricKind = iota
	CounterKind
)

type Metric struct {
	kind  MetricKind
	name  string
	value uint64
}

func NewMetricGauge(newName string, newValue Gauge) Metric {
	return Metric{
		kind:  GaugeKind,
		name:  newName,
		value: math.Float64bits(float64(newValue)),
	}
}

func NewMetricCounter(newName string, newValue Counter) Metric {
	return Metric{
		kind:  CounterKind,
		name:  newName,
		value: uint64(newValue),
	}
}

func (m *Metric) GetKind() string {
	switch m.kind {
	case GaugeKind:
		return "gauge"
	case CounterKind:
		return "counter"
	default:
		log.Println("Error: unsupported type: ", m.kind)
		return "unsupported"
	}
}

func (m *Metric) GetName() string {
	return m.name
}

func (m *Metric) GetGaugeValue() Gauge {
	return Gauge(math.Float64frombits(m.value))
}

func (m *Metric) GetCounterValue() Counter {
	return Counter(m.value)
}

type metricRepository interface {
	GetMetric(name string) (Metric, error)
	Update(Metric)
	BatchUpdate(metrics []Metric)
}

func UpdateMetrics(metrics metricRepository) {
	log.Println("reading MemStats")
	runtimeMetrics := runtime.MemStats{}
	runtime.ReadMemStats(&runtimeMetrics)

	log.Println("updating metrics")
	pollCount, _ := metrics.GetMetric("PollCount")
	metrics.BatchUpdate(
		[]Metric{
			NewMetricGauge("Alloc", Gauge(runtimeMetrics.Alloc)),
			NewMetricGauge("BuckHashSys", Gauge(runtimeMetrics.BuckHashSys)),
			NewMetricGauge("Frees", Gauge(runtimeMetrics.Frees)),
			NewMetricGauge("GCCPUFraction", Gauge(runtimeMetrics.GCCPUFraction)),
			NewMetricGauge("GCSys", Gauge(runtimeMetrics.GCSys)),
			NewMetricGauge("HeapAlloc", Gauge(runtimeMetrics.HeapAlloc)),
			NewMetricGauge("HeapIdle", Gauge(runtimeMetrics.HeapIdle)),
			NewMetricGauge("HeapInuse", Gauge(runtimeMetrics.HeapInuse)),
			NewMetricGauge("HeapObjects", Gauge(runtimeMetrics.HeapObjects)),
			NewMetricGauge("HeapReleased", Gauge(runtimeMetrics.HeapReleased)),
			NewMetricGauge("HeapSys", Gauge(runtimeMetrics.HeapSys)),
			NewMetricGauge("LastGC", Gauge(runtimeMetrics.LastGC)),
			NewMetricGauge("Lookups", Gauge(runtimeMetrics.Lookups)),
			NewMetricGauge("MCacheInuse", Gauge(runtimeMetrics.MCacheInuse)),
			NewMetricGauge("MCacheSys", Gauge(runtimeMetrics.MCacheSys)),
			NewMetricGauge("MSpanInuse", Gauge(runtimeMetrics.MSpanInuse)),
			NewMetricGauge("MSpanSys", Gauge(runtimeMetrics.MSpanSys)),
			NewMetricGauge("Mallocs", Gauge(runtimeMetrics.Mallocs)),
			NewMetricGauge("NextGC", Gauge(runtimeMetrics.NextGC)),
			NewMetricGauge("NumForcedGC", Gauge(runtimeMetrics.NumForcedGC)),
			NewMetricGauge("NumGC", Gauge(runtimeMetrics.NumGC)),
			NewMetricGauge("OtherSys", Gauge(runtimeMetrics.OtherSys)),
			NewMetricGauge("PauseTotalNs", Gauge(runtimeMetrics.PauseTotalNs)),
			NewMetricGauge("StackInuse", Gauge(runtimeMetrics.StackInuse)),
			NewMetricGauge("StackSys", Gauge(runtimeMetrics.StackSys)),
			NewMetricGauge("Sys", Gauge(runtimeMetrics.Sys)),
			NewMetricGauge("TotalAlloc", Gauge(runtimeMetrics.TotalAlloc)),
			NewMetricCounter("PollCount", pollCount.GetCounterValue()+1),
			NewMetricGauge("RandomValue", Gauge(rand.Float64()*math.MaxFloat64)),
		},
	)
}

func UpdateMetricsGopsutil(metrics metricRepository) {
	log.Println("reading gopsutil")
	m, err := mem.VirtualMemory()
	if err != nil {
		log.Println(err)
		return
	}

	percent, err := cpu.Percent(100*time.Millisecond, false)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("updating metrics")
	metrics.BatchUpdate(
		[]Metric{
			NewMetricGauge("TotalMemory", Gauge(m.Total)),
			NewMetricGauge("FreeMemory", Gauge(m.Free)),
			NewMetricGauge("CPUutilization1", Gauge(percent[0])),
		},
	)
}

func ResetPollCounter(metrics metricRepository) {
	metrics.Update(NewMetricCounter("PollCount", 0))
}
