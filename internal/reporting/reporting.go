package reporting

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var HostUnreachableCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "host_unreachable_total",
		Help: "Number of times a host was unreachable",
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
	prometheus.MustRegister(HostUnreachableCounter)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":3009", nil))
}
