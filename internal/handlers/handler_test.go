package handlers

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateStorageHandler(t *testing.T) {
	type args struct {
		storage metricRepository
	}

	type want struct {
		mtrcs      map[string]metrics.Metric
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
				storage: repository.NewMemStorage(),
			},
			target: "/update",
			want: want{
				mtrcs:      map[string]metrics.Metric{},
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "GaugeBadRequest",
			args: args{
				storage: repository.NewMemStorage(),
			},
			target: "/update/gauge/test/value",
			want: want{
				mtrcs:      map[string]metrics.Metric{},
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "CounterBadRequest",
			args: args{
				storage: repository.NewMemStorage(),
			},
			target: "/update/counter/test/value",
			want: want{
				mtrcs:      map[string]metrics.Metric{},
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "NotImplemented",
			args: args{
				storage: repository.NewMemStorage(),
			},
			target: "/update/something/test/value",
			want: want{
				mtrcs:      map[string]metrics.Metric{},
				statusCode: http.StatusNotImplemented,
			},
		},
		{
			name: "GaugeOK",
			args: args{
				storage: repository.NewMemStorage(),
			},
			target: "/update/gauge/test/12",
			want: want{
				mtrcs: map[string]metrics.Metric{
					"test": metrics.NewMetricGauge("test", 12),
				},
				statusCode: http.StatusOK,
			},
		},
		{
			name: "CounterOK",
			args: args{
				storage: repository.NewMemStorage(),
			},
			target: "/update/counter/test/12",
			want: want{
				mtrcs: map[string]metrics.Metric{
					"test": metrics.NewMetricCounter("test", 12),
				},
				statusCode: http.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Post("/update/{kind}/{name}/{value}", UpdateStorageHandler(tt.args.storage, ""))

			request := httptest.NewRequest(http.MethodPost, tt.target, nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)
			result := recorder.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				assert.Equal(t, tt.want.mtrcs, tt.args.storage.GetMetricsMap())
			}
		})
	}
}

func TestPrintStorageHandler(t *testing.T) {
	storage := repository.NewMemStorage()
	storage.Update(metrics.NewMetricCounter("testC", 123))
	storage.Update(metrics.NewMetricGauge("testG", 123))
	storage.Update(metrics.NewMetricCounter("testC", 321))
	storage.Update(metrics.NewMetricGauge("testG", 321))

	type want struct {
		statusCode int
		html       string
	}

	tests := []struct {
		name    string
		storage metricRepository
		want    want
	}{
		{
			name:    "OkStorage",
			storage: storage,
			want: want{
				statusCode: http.StatusOK,
				html:       "counter testC 444gauge testG 321.000",
			},
		},
		{
			name:    "EmptyStorage",
			storage: repository.NewMemStorage(),
			want: want{
				statusCode: http.StatusOK,
				html:       "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Get("/", PrintStorageHandler(tt.storage))

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)
			result := recorder.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				slice, _ := io.ReadAll(result.Body)
				str := string(slice[:])
				assert.Equal(t, tt.want.html, str)
			}
		})
	}
}

func TestPrintValueHandler(t *testing.T) {
	storage := repository.NewMemStorage()
	storage.Update(metrics.NewMetricCounter("testC", 123))
	storage.Update(metrics.NewMetricGauge("testG", 123))
	storage.Update(metrics.NewMetricCounter("testC", 321))
	storage.Update(metrics.NewMetricGauge("testG", 321))

	type want struct {
		statusCode int
		html       string
	}

	tests := []struct {
		name    string
		target  string
		storage metricRepository
		want    want
	}{
		{
			name:    "CounterOk",
			target:  "/value/counter/testC",
			storage: storage,
			want: want{
				statusCode: http.StatusOK,
				html:       "444",
			},
		},
		{
			name:    "GaugeOk",
			target:  "/value/gauge/testG",
			storage: storage,
			want: want{
				statusCode: http.StatusOK,
				html:       "321.000",
			},
		},
		{
			name:    "NotFound",
			target:  "/value/counter/Polls",
			storage: storage,
			want: want{
				statusCode: http.StatusNotFound,
				html:       "",
			},
		},
		{
			name:    "EmptyStorage",
			target:  "/value/counter/Polls",
			storage: repository.NewMemStorage(),
			want: want{
				statusCode: http.StatusNotFound,
				html:       "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Get("/value/{kind}/{name}", PrintValueHandler(tt.storage, ""))

			request := httptest.NewRequest(http.MethodGet, tt.target, nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)
			result := recorder.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				slice, _ := io.ReadAll(result.Body)
				str := string(slice[:])
				assert.Equal(t, tt.want.html, str)
			}
		})
	}
}

func TestNewJSONMetric(t *testing.T) {
	delta := int64(12)
	counter := metrics.NewMetricCounter("test", metrics.Counter(delta))
	counterJSON := &JSONMetric{ID: "test", Delta: &delta}

	value := 12.21
	gauge := metrics.NewMetricGauge("test", metrics.Gauge(value))
	gaugeJSON := &JSONMetric{ID: "test", Value: &value}

	tests := []struct {
		name   string
		metric metrics.Metric
		want   *JSONMetric
	}{
		{
			name:   "Counter",
			metric: counter,
			want:   counterJSON,
		},
		{
			name:   "Gauge",
			metric: gauge,
			want:   gaugeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonObj, err := NewJSONMetric(tt.metric)
			require.NoError(t, err)
			require.Equal(t, tt.want.ID, jsonObj.ID)

			if tt.want.Delta != nil {
				require.Equal(t, *tt.want.Delta, *jsonObj.Delta)
			}

			if tt.want.Value != nil {
				require.Equal(t, *tt.want.Value, *jsonObj.Value)
			}
		})
	}
}
