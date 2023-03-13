package clients

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/handlers"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/metrics"
	"github.com/VladimirMovsesyan/praktikum-devops/internal/repository"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestMetricsUpload(t *testing.T) {
	client := NewMetricsClient()
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
			storage := &repository.MemStorage{}
			server := httptest.NewServer(handlers.UpdateStorageHandler(storage))
			defer server.Close()
			mtrcs := metrics.NewMetrics()
			metrics.UpdateMetrics(mtrcs)
			MetricsUpload(client, mtrcs, server.URL)
			require.Equal(t, mtrcs.MetricSlice, storage.GetMetrics())
		})
	}
}
