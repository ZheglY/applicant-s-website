package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsHandler(t *testing.T) {
	metrics := NewMetrics()
	metrics.HTTPRequests.WithLabelValues(http.MethodGet, "GET /health/live", "200").Inc()
	recorder := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "unik_http_requests_total") {
		t.Fatal("custom HTTP metric is missing")
	}
}
