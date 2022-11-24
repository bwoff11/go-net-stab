package registry

import (
	"log"
	"net"
	"time"

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

	// Remove ping from outstanding pings
	var removed bool
	for i, p := range pending {
		if p.PingerID == ping.PingerID && p.Sequence == ping.Sequence {
			pending = append(pending[:i], pending[i+1:]...)
			removed = true
			break
		}
	}
	if !removed {
		log.Println("Failed to remove ping from pending pings")
	}
}

func HandleEchoReply(pingerID int, sequence int, host net.Addr) {
	// TODO: Add check for reply to timed-out packet

	// Search pending pings for a match
	for _, ping := range pending {
		if ping.PingerID == pingerID && ping.Sequence == sequence {
			handlePingMatch(ping)
			return
		}
	}
}

func handlePingMatch(ping Ping) {
	now := time.Now()
	ping.ReceivedAt = &now
	rtt := float64(ping.CalculateRoundTripTime().Milliseconds())

	reporting.RttGauge.With(
		prometheus.Labels{
			"source_ip":      ping.SourceIP,
			"destination_ip": ping.DestinationIP,
		},
	).Set(rtt)
	log.Println("RTT:", rtt)

	// Remove ping from outstanding pings
	var removed bool
	for i, p := range pending {
		if p.PingerID == ping.PingerID && p.Sequence == ping.Sequence {
			pending = append(pending[:i], pending[i+1:]...)
			removed = true
			break
		}
	}
	if !removed {
		log.Println("Failed to remove ping from pending pings")
	}
}
