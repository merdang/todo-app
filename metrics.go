package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusMetric struct {
	uriCounter            prometheus.CounterVec
	responseTimeHistogram prometheus.HistogramVec
}

func NewPrometheusMetric() (*PrometheusMetric, error) {
	return &PrometheusMetric{
		uriCounter: *prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "uri", "status"},
		),
		responseTimeHistogram: *prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Histogram of response latency (seconds).",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "uri"},
		),
	}, nil
}

func (s *PrometheusMetric) Init() {
	prometheus.MustRegister(s.uriCounter)
	prometheus.MustRegister(s.responseTimeHistogram)
}
