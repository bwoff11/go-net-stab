package registry

import (
	"log"
	"net"

	"github.com/bwoff11/go-net-stab/internal/reporting"
	"github.com/prometheus/client_golang/prometheus"
)

func handleLostPing(ping Ping) {
	log.Println("Lost ping:", ping)
	reporting.LostPacketCounter.With(
		prometheus.Labels{
			"source_ip":      ping.SourceIP,
			"destination_ip": ping.DestinationIP,
		},
	).Inc()

	// TODO: Remove ping from outstanding pings
}

func HandleEchoReply(pingerID int, sequence int, host net.Addr) {
	// TODO: Add check for reply to timed-out packet

	// TODO: Search pending pings for a match
}

func handlePingMatch(ping Ping) {
	ping.SetReceived()

	reporting.RttGauge.With(
		prometheus.Labels{
			"source_ip":      ping.SourceIP,
			"destination_ip": ping.DestinationIP,
		},
	).Set(float64(ping.CalculateRoundTripTime().Milliseconds()))

	// TODO: Remove ping from outstanding pings
}
