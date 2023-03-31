package repository

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage_GetMetrics(t *testing.T) {
	type fields struct {
		mtrcs map[string]metrics.Metric
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]metrics.Metric
	}{
		{
			name: "Simple test",
			fields: fields{
				mtrcs: map[string]metrics.Metric{
					"firstC":  metrics.NewMetricCounter("firstC", 1),
					"secondC": metrics.NewMetricCounter("secondC", 2),
					"thirdC":  metrics.NewMetricCounter("thirdC", 3),
					"firstG":  metrics.NewMetricGauge("firstG", 1),
					"secondG": metrics.NewMetricGauge("secondG", 2),
					"thirdG":  metrics.NewMetricGauge("thirdG", 3),
				},
			},
			want: map[string]metrics.Metric{
				"firstC":  metrics.NewMetricCounter("firstC", 1),
				"secondC": metrics.NewMetricCounter("secondC", 2),
				"thirdC":  metrics.NewMetricCounter("thirdC", 3),
				"firstG":  metrics.NewMetricGauge("firstG", 1),
				"secondG": metrics.NewMetricGauge("secondG", 2),
				"thirdG":  metrics.NewMetricGauge("thirdG", 3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				mtrcs: tt.fields.mtrcs,
			}
			assert.Equal(t, tt.want, ms.GetMetricsMap())
		})
	}
}

func TestMemStorage_Update(t *testing.T) {
	type fields struct {
		mtrcs map[string]metrics.Metric
	}

	type args struct {
		newMetric metrics.Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			name: "Normal counter update",
			fields: fields{
				mtrcs: map[string]metrics.Metric{
					"firstC":  metrics.NewMetricCounter("firstC", 1),
					"secondC": metrics.NewMetricCounter("secondC", 2),
					"thirdC":  metrics.NewMetricCounter("thirdC", 3),
					"firstG":  metrics.NewMetricGauge("firstG", 1),
					"secondG": metrics.NewMetricGauge("secondG", 2),
					"thirdG":  metrics.NewMetricGauge("thirdG", 3),
				},
			},
			args: args{
				newMetric: metrics.NewMetricCounter("firstC", 100),
			},
			want: fields{
				mtrcs: map[string]metrics.Metric{
					"firstC":  metrics.NewMetricCounter("firstC", 101),
					"secondC": metrics.NewMetricCounter("secondC", 2),
					"thirdC":  metrics.NewMetricCounter("thirdC", 3),
					"firstG":  metrics.NewMetricGauge("firstG", 1),
					"secondG": metrics.NewMetricGauge("secondG", 2),
					"thirdG":  metrics.NewMetricGauge("thirdG", 3),
				},
			},
		},
		{
			name: "Normal gauge update",
			fields: fields{
				mtrcs: map[string]metrics.Metric{
					"firstC":  metrics.NewMetricCounter("firstC", 1),
					"secondC": metrics.NewMetricCounter("secondC", 2),
					"thirdC":  metrics.NewMetricCounter("thirdC", 3),
					"firstG":  metrics.NewMetricGauge("firstG", 1),
					"secondG": metrics.NewMetricGauge("secondG", 2),
					"thirdG":  metrics.NewMetricGauge("thirdG", 3),
				},
			},
			args: args{
				newMetric: metrics.NewMetricGauge("firstG", 100),
			},
			want: fields{
				mtrcs: map[string]metrics.Metric{
					"firstC":  metrics.NewMetricCounter("firstC", 1),
					"secondC": metrics.NewMetricCounter("secondC", 2),
					"thirdC":  metrics.NewMetricCounter("thirdC", 3),
					"firstG":  metrics.NewMetricGauge("firstG", 100),
					"secondG": metrics.NewMetricGauge("secondG", 2),
					"thirdG":  metrics.NewMetricGauge("thirdG", 3),
				},
			},
		},
		{
			name: "New metric counter update",
			fields: fields{
				mtrcs: map[string]metrics.Metric{
					"firstC":  metrics.NewMetricCounter("firstC", 1),
					"secondC": metrics.NewMetricCounter("secondC", 2),
					"thirdC":  metrics.NewMetricCounter("thirdC", 3),
					"firstG":  metrics.NewMetricGauge("firstG", 1),
					"secondG": metrics.NewMetricGauge("secondG", 2),
					"thirdG":  metrics.NewMetricGauge("thirdG", 3),
				},
			},
			args: args{
				newMetric: metrics.NewMetricCounter("fourthG", 4),
			},
			want: fields{
				mtrcs: map[string]metrics.Metric{
					"firstC":  metrics.NewMetricCounter("firstC", 1),
					"secondC": metrics.NewMetricCounter("secondC", 2),
					"thirdC":  metrics.NewMetricCounter("thirdC", 3),
					"firstG":  metrics.NewMetricGauge("firstG", 1),
					"secondG": metrics.NewMetricGauge("secondG", 2),
					"thirdG":  metrics.NewMetricGauge("thirdG", 3),
					"fourthG": metrics.NewMetricCounter("fourthG", 4),
				},
			},
		},
		{
			name: "New metric gauge update",
			fields: fields{
				mtrcs: map[string]metrics.Metric{
					"firstC":  metrics.NewMetricCounter("firstC", 1),
					"secondC": metrics.NewMetricCounter("secondC", 2),
					"thirdC":  metrics.NewMetricCounter("thirdC", 3),
					"firstG":  metrics.NewMetricGauge("firstG", 1),
					"secondG": metrics.NewMetricGauge("secondG", 2),
					"thirdG":  metrics.NewMetricGauge("thirdG", 3),
				},
			},
			args: args{
				newMetric: metrics.NewMetricGauge("fourthG", 4),
			},
			want: fields{
				mtrcs: map[string]metrics.Metric{
					"firstC":  metrics.NewMetricCounter("firstC", 1),
					"secondC": metrics.NewMetricCounter("secondC", 2),
					"thirdC":  metrics.NewMetricCounter("thirdC", 3),
					"firstG":  metrics.NewMetricGauge("firstG", 1),
					"secondG": metrics.NewMetricGauge("secondG", 2),
					"thirdG":  metrics.NewMetricGauge("thirdG", 3),
					"fourthG": metrics.NewMetricGauge("fourthG", 4),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				mtrcs: tt.fields.mtrcs,
			}
			ms.Update(tt.args.newMetric)
			assert.Equal(t, tt.want.mtrcs, ms.mtrcs)
		})
	}
}
