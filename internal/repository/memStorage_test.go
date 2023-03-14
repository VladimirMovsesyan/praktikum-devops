package repository

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage_GetMetrics(t *testing.T) {
	type fields struct {
		mtrcs []metrics.Metric
	}
	tests := []struct {
		name   string
		fields fields
		want   []metrics.Metric
	}{
		{
			name: "Simple test",
			fields: fields{
				mtrcs: []metrics.Metric{
					metrics.NewMetricCounter("first", 1),
					metrics.NewMetricCounter("second", 2),
					metrics.NewMetricCounter("third", 3),
					metrics.NewMetricGauge("first", 1),
					metrics.NewMetricGauge("second", 2),
					metrics.NewMetricGauge("third", 3),
				},
			},
			want: []metrics.Metric{
				metrics.NewMetricCounter("first", 1),
				metrics.NewMetricCounter("second", 2),
				metrics.NewMetricCounter("third", 3),
				metrics.NewMetricGauge("first", 1),
				metrics.NewMetricGauge("second", 2),
				metrics.NewMetricGauge("third", 3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				mtrcs: tt.fields.mtrcs,
			}
			assert.Equal(t, tt.want, ms.GetMetrics())
		})
	}
}

func TestMemStorage_Update(t *testing.T) {
	type fields struct {
		mtrcs []metrics.Metric
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
				mtrcs: []metrics.Metric{
					metrics.NewMetricCounter("first", 1),
					metrics.NewMetricCounter("second", 2),
					metrics.NewMetricCounter("third", 3),
					metrics.NewMetricGauge("first", 1),
					metrics.NewMetricGauge("second", 2),
					metrics.NewMetricGauge("third", 3),
				},
			},
			args: args{
				newMetric: metrics.NewMetricCounter("first", 100),
			},
			want: fields{
				mtrcs: []metrics.Metric{
					metrics.NewMetricCounter("first", 101),
					metrics.NewMetricCounter("second", 2),
					metrics.NewMetricCounter("third", 3),
					metrics.NewMetricGauge("first", 1),
					metrics.NewMetricGauge("second", 2),
					metrics.NewMetricGauge("third", 3),
				},
			},
		},
		{
			name: "Normal gauge update",
			fields: fields{
				mtrcs: []metrics.Metric{
					metrics.NewMetricCounter("first", 1),
					metrics.NewMetricCounter("second", 2),
					metrics.NewMetricCounter("third", 3),
					metrics.NewMetricGauge("first", 1),
					metrics.NewMetricGauge("second", 2),
					metrics.NewMetricGauge("third", 3),
				},
			},
			args: args{
				newMetric: metrics.NewMetricGauge("first", 100),
			},
			want: fields{
				mtrcs: []metrics.Metric{
					metrics.NewMetricCounter("first", 1),
					metrics.NewMetricCounter("second", 2),
					metrics.NewMetricCounter("third", 3),
					metrics.NewMetricGauge("first", 100),
					metrics.NewMetricGauge("second", 2),
					metrics.NewMetricGauge("third", 3),
				},
			},
		},
		{
			name: "New metric counter update",
			fields: fields{
				mtrcs: []metrics.Metric{
					metrics.NewMetricCounter("first", 1),
					metrics.NewMetricCounter("second", 2),
					metrics.NewMetricCounter("third", 3),
					metrics.NewMetricGauge("first", 1),
					metrics.NewMetricGauge("second", 2),
					metrics.NewMetricGauge("third", 3),
				},
			},
			args: args{
				newMetric: metrics.NewMetricCounter("fourth", 4),
			},
			want: fields{
				mtrcs: []metrics.Metric{
					metrics.NewMetricCounter("first", 1),
					metrics.NewMetricCounter("second", 2),
					metrics.NewMetricCounter("third", 3),
					metrics.NewMetricGauge("first", 1),
					metrics.NewMetricGauge("second", 2),
					metrics.NewMetricGauge("third", 3),
					metrics.NewMetricCounter("fourth", 4),
				},
			},
		},
		{
			name: "New metric gauge update",
			fields: fields{
				mtrcs: []metrics.Metric{
					metrics.NewMetricCounter("first", 1),
					metrics.NewMetricCounter("second", 2),
					metrics.NewMetricCounter("third", 3),
					metrics.NewMetricGauge("first", 1),
					metrics.NewMetricGauge("second", 2),
					metrics.NewMetricGauge("third", 3),
				},
			},
			args: args{
				newMetric: metrics.NewMetricGauge("fourth", 4),
			},
			want: fields{
				mtrcs: []metrics.Metric{
					metrics.NewMetricCounter("first", 1),
					metrics.NewMetricCounter("second", 2),
					metrics.NewMetricCounter("third", 3),
					metrics.NewMetricGauge("first", 1),
					metrics.NewMetricGauge("second", 2),
					metrics.NewMetricGauge("third", 3),
					metrics.NewMetricGauge("fourth", 4),
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

func Test_sameMetric(t *testing.T) {
	type args struct {
		metric1 metrics.Metric
		metric2 metrics.Metric
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Same",
			args: args{
				metric1: metrics.NewMetricCounter("test", 1),
				metric2: metrics.NewMetricCounter("test", 1),
			},
			want: true,
		},
		{
			name: "name diff",
			args: args{
				metric1: metrics.NewMetricCounter("tost", 1),
				metric2: metrics.NewMetricCounter("test", 1),
			},
			want: false,
		},
		{
			name: "kind diff",
			args: args{
				metric1: metrics.NewMetricCounter("test", 1),
				metric2: metrics.NewMetricGauge("test", 1),
			},
			want: false,
		},
		{
			name: "name and kind diff",
			args: args{
				metric1: metrics.NewMetricCounter("tost", 1),
				metric2: metrics.NewMetricGauge("test", 1),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sameMetric(tt.args.metric1, tt.args.metric2))
		})
	}
}
