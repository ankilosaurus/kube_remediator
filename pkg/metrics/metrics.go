package metrics

import (
	httpmux "github.com/google/cadvisor/http/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics interface {
	RegisterMetrics()
}

func RegisterHandler(mux httpmux.Mux) error {
	mux.Handle("/metrics", promhttp.Handler())
	return nil
}
