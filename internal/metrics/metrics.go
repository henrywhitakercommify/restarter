package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	TotalPods = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "restarter_running_pods",
		},
		[]string{"namespace", "deployment"},
	)
	ReadyPods = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "restarter_ready_pods",
		},
		[]string{"namespace", "deployment"},
	)
	TotalRestarts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "restarter_restarts_total",
		},
		[]string{"namespace", "deployment"},
	)
	TotalChecks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "restarter_checks_total",
		},
		[]string{"namespace", "deployment"},
	)
)

func init() {
	prometheus.DefaultRegisterer.MustRegister(TotalPods)
	prometheus.DefaultRegisterer.MustRegister(ReadyPods)
	prometheus.DefaultRegisterer.MustRegister(TotalRestarts)
	prometheus.DefaultRegisterer.MustRegister(TotalChecks)
}

func Server(port int) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	return server
}
