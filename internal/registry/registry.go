package registry

import (
	"log"
	"net"
	"time"

	"github.com/bwoff11/go-net-stab/internal/reporting"
	"github.com/prometheus/client_golang/prometheus"
)

var SentPings chan Ping
var outstandingPings []Ping

func Start() error {

	// Monitor the sentPings channel and import any new pings into the outstandingPings slice
	SentPings = make(chan Ping)
	go func() {
		for {
			ping := <-SentPings
			outstandingPings = append(outstandingPings, ping)
		}
	}()
	log.Println("Registry successfully started and waiting for pings")
	return nil
}

func HandleEchoReply(pingerID int, sequence int, host net.Addr) {
	// TODO: Add check for reply to timed-out packet

	// Search outstanding pings for a match
	for _, ping := range outstandingPings {
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
	for i, p := range outstandingPings {
		if p == ping {
			outstandingPings = append(outstandingPings[:i], outstandingPings[i+1:]...)
			return
		}
	}
}
