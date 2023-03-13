package handlers

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateStorageHandler(t *testing.T) {
	type args struct {
		storage repository.MetricRepository
	}

	type want struct {
		mtrcs      []metrics.Metric
		statusCode int
	}

	tests := []struct {
		name   string
		args   args
		target string
		want   want
	}{
		{
			name: "NotFound",
			args: args{
				storage: &repository.MemStorage{},
			},
			target: "/update",
			want: want{
				mtrcs:      []metrics.Metric{},
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "GaugeBadRequest",
			args: args{
				storage: &repository.MemStorage{},
			},
			target: "/update/gauge/test/value",
			want: want{
				mtrcs:      []metrics.Metric{},
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "CounterBadRequest",
			args: args{
				storage: &repository.MemStorage{},
			},
			target: "/update/counter/test/value",
			want: want{
				mtrcs:      []metrics.Metric{},
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "NotImplemented",
			args: args{
				storage: &repository.MemStorage{},
			},
			target: "/update/something/test/value",
			want: want{
				mtrcs:      []metrics.Metric{},
				statusCode: http.StatusNotImplemented,
			},
		},
		{
			name: "GaugeOK",
			args: args{
				storage: &repository.MemStorage{},
			},
			target: "/update/gauge/test/12",
			want: want{
				mtrcs: []metrics.Metric{
					metrics.NewMetricGauge("test", 12),
				},
				statusCode: http.StatusOK,
			},
		},
		{
			name: "CounterOK",
			args: args{
				storage: &repository.MemStorage{},
			},
			target: "/update/counter/test/12",
			want: want{
				mtrcs: []metrics.Metric{
					metrics.NewMetricCounter("test", 12),
				},
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.target, nil)
			rec := httptest.NewRecorder()
			handler := UpdateStorageHandler(tt.args.storage)
			handler(rec, request)
			result := rec.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				assert.Equal(t, tt.want.mtrcs, tt.args.storage.GetMetrics())
			}
		})
	}
}
