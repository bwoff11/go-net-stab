package reporting

import (
	"log"
	"net/http"

	"github.com/bwoff11/go-net-stab/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var SentPingsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ping_sent_packet_total",
		Help: "Total number of packets sent",
	},
	[]string{"source_ip", "destination_ip"},
)

var LostPingsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ping_lost_packet_total",
		Help: "Total number of packets lost",
	},
	[]string{"source_ip", "destination_ip"},
)

var RttGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "ping_rtt_milliseconds",
		Help: "Round trip time in milliseconds",
	},
	[]string{"source_ip", "destination_ip"},
)

func ServeMetrics() {
	prometheus.MustRegister(SentPingsCounter)
	prometheus.MustRegister(LostPingsCounter)
	prometheus.MustRegister(RttGauge)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+config.Config.Port, nil))
}
