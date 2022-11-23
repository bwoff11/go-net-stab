package reporting

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":3009", nil))
}
