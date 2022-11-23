package registry

import (
	"log"
	"time"

	"github.com/bwoff11/go-net-stab/internal/reporting"
	"github.com/prometheus/client_golang/prometheus"
)

var SentPings chan Ping
var pending []Ping

func Start() error {

	startPendingImporter()
	startLostPingChecker()

	log.Println("Registry successfully started and is now waiting for pings from pingers")
	return nil
}

func startPendingImporter() {
	SentPings = make(chan Ping)
	go func() {
		for {
			ping := <-SentPings
			pending = append(pending, ping)
		}
	}()
}

func startLostPingChecker() {
	go func() {
		for {
			for _, ping := range pending {
				if ping.IsLost() {
					handleLostPing(ping)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

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
