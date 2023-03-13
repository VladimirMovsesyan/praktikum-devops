package handlers

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
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
			router := chi.NewRouter()
			router.Post("/update/{kind}/{name}/{value}", UpdateStorageHandler(tt.args.storage))

			request := httptest.NewRequest(http.MethodPost, tt.target, nil)
			recorder := httptest.NewRecorder()
			
			router.ServeHTTP(recorder, request)
			result := recorder.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				assert.Equal(t, tt.want.mtrcs, tt.args.storage.GetMetrics())
			}

			err := result.Body.Close()
			if err != nil {
				log.Println("Error: ", err)
			}
		})
	}
}

func TestPrintStorageHandler(t *testing.T) {
	storage := &repository.MemStorage{}
	storage.Update(metrics.NewMetricCounter("test", 123))
	storage.Update(metrics.NewMetricGauge("test", 123))
	storage.Update(metrics.NewMetricCounter("test", 321))
	storage.Update(metrics.NewMetricGauge("test", 321))

	type want struct {
		statusCode int
		html       string
	}

	tests := []struct {
		name    string
		storage repository.MetricRepository
		want    want
	}{
		{
			name:    "OkStorage",
			storage: storage,
			want: want{
				statusCode: http.StatusOK,
				html:       "<h1>counter test 444</h1><h1>gauge test 321.000000</h1>",
			},
		},
		{
			name:    "EmptyStorage",
			storage: &repository.MemStorage{},
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

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				slice, _ := io.ReadAll(result.Body)
				str := string(slice[:])
				assert.Equal(t, tt.want.html, str)
			}

			err := result.Body.Close()
			if err != nil {
				log.Println("Error: ", err)
			}
		})
	}
}

func TestPrintValueHandler(t *testing.T) {
	storage := &repository.MemStorage{}
	storage.Update(metrics.NewMetricCounter("test", 123))
	storage.Update(metrics.NewMetricGauge("test", 123))
	storage.Update(metrics.NewMetricCounter("test", 321))
	storage.Update(metrics.NewMetricGauge("test", 321))

	type want struct {
		statusCode int
		html       string
	}

	tests := []struct {
		name    string
		target  string
		storage repository.MetricRepository
		want    want
	}{
		{
			name:    "CounterOk",
			target:  "/value/counter/test",
			storage: storage,
			want: want{
				statusCode: http.StatusOK,
				html:       "<h1>counter test 444</h1>",
			},
		},
		{
			name:    "GaugeOk",
			target:  "/value/gauge/test",
			storage: storage,
			want: want{
				statusCode: http.StatusOK,
				html:       "<h1>gauge test 321.000000</h1>",
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
			storage: &repository.MemStorage{},
			want: want{
				statusCode: http.StatusNotFound,
				html:       "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Get("/value/{kind}/{name}", PrintValueHandler(tt.storage))

			request := httptest.NewRequest(http.MethodGet, tt.target, nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)
			result := recorder.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				slice, _ := io.ReadAll(result.Body)
				str := string(slice[:])
				assert.Equal(t, tt.want.html, str)
			}

			err := result.Body.Close()
			if err != nil {
				log.Println("Error: ", err)
			}
		})
	}
}
