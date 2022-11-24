package reporting

import (
	"log"
	"net/http"

	"github.com/bwoff11/go-net-stab/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var SentPacketCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ping_sent_packet_total",
		Help: "Total number of packets sent",
	},
	[]string{"source_ip", "destination_ip"},
)

var DestinationUnreachableCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ping_destination_unreachable_total",
		Help: "Total number of destination unreachable packets received",
	},
	[]string{"source_ip", "destination_ip"},
)

var TimeExceededCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ping_time_exceeded_total",
		Help: "Total number of time exceeded packets received",
	},
	[]string{"source_ip", "destination_ip"},
)

var LostPacketCounter = prometheus.NewCounterVec(
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
	prometheus.MustRegister(SentPacketCounter)
	prometheus.MustRegister(DestinationUnreachableCounter)
	prometheus.MustRegister(TimeExceededCounter)
	prometheus.MustRegister(LostPacketCounter)
	prometheus.MustRegister(RttGauge)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+config.Config.Port, nil))
}
