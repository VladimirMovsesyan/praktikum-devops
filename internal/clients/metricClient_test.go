package clients

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/handlers"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestMetricsUpload(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Test #1",
		},
		{
			name: "Test #2",
		},
		{
			name: "Test #3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewMemStorage()
			router := chi.NewRouter()
			router.Post("/update/{kind}/{name}/{value}", handlers.UpdateStorageHandler(storage))

			server := httptest.NewServer(router)
			defer server.Close()

			mtrcs := metrics.NewMetrics()
			metrics.UpdateMetrics(mtrcs)

			for _, metric := range mtrcs.MetricSlice {
				metricUpload(server.URL, metric)
			}

			storageMetrics := storage.GetMetrics()
			for _, value := range mtrcs.MetricSlice {
				require.Equal(t, value, storageMetrics[value.GetName()])
			}
		})
	}
}
