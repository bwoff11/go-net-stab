package reporting

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var SendPacketCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "send_packet_total",
		Help: "Total number of packets sent",
	},
	[]string{"source_ip", "destination_ip"},
)

var RttGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "rtt",
		Help: "Round trip time in milliseconds",
	},
	[]string{"source_ip", "destination_ip"},
)

func ServeMetrics() {
	prometheus.MustRegister(RttGauge)
	prometheus.MustRegister(SendPacketCounter)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":3009", nil))
}
