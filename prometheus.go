package main

import "github.com/prometheus/client_golang/prometheus"

var packetLossGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "packet_loss",
		Help: "Packet loss percentage",
	},
	[]string{"source_ip", "destination_ip"},
)

var rttGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "rtt",
		Help: "Round trip time in milliseconds",
	},
	[]string{"source_ip", "destination_ip"},
)

func registerPrometheusMetrics() {
	prometheus.MustRegister(packetLossGauge)
	prometheus.MustRegister(rttGauge)
}
