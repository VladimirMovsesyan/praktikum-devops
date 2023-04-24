package metrics

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestMetric_GetCounterValue(t *testing.T) {
	type fields struct {
		kind  MetricKind
		name  string
		value uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   Counter
	}{
		{
			name: "Zero value",
			fields: fields{
				kind:  CounterKind,
				name:  "name",
				value: 0,
			},
			want: Counter(0),
		},
		{
			name: "Small, but not zero value",
			fields: fields{
				kind:  CounterKind,
				name:  "name",
				value: 5,
			},
			want: Counter(5),
		},
		{
			name: "Large value",
			fields: fields{
				kind:  CounterKind,
				name:  "name",
				value: math.MaxInt64,
			},
			want: Counter(math.MaxInt64),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				kind:  tt.fields.kind,
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			assert.Equal(t, tt.want, m.GetCounterValue())
		})
	}
}

func TestMetric_GetGaugeValue(t *testing.T) {
	type fields struct {
		kind  MetricKind
		name  string
		value uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   Gauge
	}{
		{
			name: "Zero value",
			fields: fields{
				kind:  GaugeKind,
				name:  "name",
				value: math.Float64bits(0),
			},
			want: Gauge(0),
		},
		{
			name: "Small, but not zero value",
			fields: fields{
				kind:  GaugeKind,
				name:  "name",
				value: math.Float64bits(5),
			},
			want: Gauge(5),
		},
		{
			name: "Large value",
			fields: fields{
				kind:  GaugeKind,
				name:  "name",
				value: math.Float64bits(math.MaxFloat64),
			},
			want: Gauge(math.MaxFloat64),
		},
		{
			name: "Negative large value",
			fields: fields{
				kind:  GaugeKind,
				name:  "name",
				value: math.Float64bits(-math.MaxFloat64),
			},
			want: Gauge(-math.MaxFloat64),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				kind:  tt.fields.kind,
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			assert.Equal(t, tt.want, m.GetGaugeValue())
		})
	}
}

func TestMetric_GetKind(t *testing.T) {
	type fields struct {
		kind  MetricKind
		name  string
		value uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Zero value",
			fields: fields{
				kind:  CounterKind,
				name:  "name",
				value: 0,
			},
			want: "counter",
		},
		{
			name: "Small, but not zero value",
			fields: fields{
				kind:  CounterKind,
				name:  "name",
				value: 5,
			},
			want: "counter",
		},
		{
			name: "Large value",
			fields: fields{
				kind:  GaugeKind,
				name:  "name",
				value: math.Float64bits(math.MaxFloat64),
			},
			want: "gauge",
		},
		{
			name: "Negative large value",
			fields: fields{
				kind:  GaugeKind,
				name:  "name",
				value: math.Float64bits(-math.MaxFloat64),
			},
			want: "gauge",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				kind:  tt.fields.kind,
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			assert.Equal(t, tt.want, m.GetKind())
		})
	}
}

func TestMetric_GetName(t *testing.T) {
	type fields struct {
		kind  MetricKind
		name  string
		value uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Zero value",
			fields: fields{
				kind:  CounterKind,
				name:  "1",
				value: 0,
			},
			want: "1",
		},
		{
			name: "Small, but not zero value",
			fields: fields{
				kind:  CounterKind,
				name:  "2",
				value: 5,
			},
			want: "2",
		},
		{
			name: "Large value",
			fields: fields{
				kind:  GaugeKind,
				name:  "3",
				value: math.Float64bits(math.MaxFloat64),
			},
			want: "3",
		},
		{
			name: "Negative large value",
			fields: fields{
				kind:  GaugeKind,
				name:  "4",
				value: math.Float64bits(-math.MaxFloat64),
			},
			want: "4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				kind:  tt.fields.kind,
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			assert.Equal(t, tt.want, m.GetName())
		})
	}
}
