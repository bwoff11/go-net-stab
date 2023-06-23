package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics structure holds all the Prometheus metric counters and gauges
type Metrics struct {
	SentPingsCounter *prometheus.CounterVec
	LostPingsCounter *prometheus.CounterVec
	RttGauge         *prometheus.GaugeVec
}

// NewMetrics initializes and returns a new Metrics structure
func New() *Metrics {
	// Initialize Prometheus metrics
	return &Metrics{
		SentPingsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ping_sent_packet_total",
				Help: "Total number of packets sent",
			},
			[]string{
				"source_hostname",
				"destination_hostname",
				"destination_address",
				"destination_location",
			},
		),

		LostPingsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ping_lost_packet_total",
				Help: "Total number of packets lost",
			},
			[]string{
				"source_hostname",
				"destination_hostname",
				"destination_address",
				"destination_location",
			},
		),

		RttGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ping_rtt_milliseconds",
				Help: "Round trip time in milliseconds",
			},
			[]string{
				"source_hostname",
				"destination_hostname",
				"destination_address",
				"destination_location",
			},
		),
	}
}

// RegisterMetrics registers all the Prometheus metrics with the Prometheus default registry
func (m *Metrics) RegisterMetrics() {
	prometheus.MustRegister(m.SentPingsCounter)
	prometheus.MustRegister(m.LostPingsCounter)
	prometheus.MustRegister(m.RttGauge)
}
