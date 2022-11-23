package main

import (
	"log"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var sentPings chan Ping
var outstandingPings []Ping

func createSentPingsChannel() {
	sentPings = make(chan Ping)
	go func() {
		for {
			ping := <-sentPings
			outstandingPings = append(outstandingPings, ping)
		}
	}()
}

func startResponseHandler() {
	connection, err := icmp.ListenPacket("ip4:icmp", "192.168.1.11")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			buffer := make([]byte, 1500)
			_, host, err := connection.ReadFrom(buffer)
			if err != nil {
				panic(err)
			}

			message, err := icmp.ParseMessage(1, buffer)
			if err != nil {
				panic(err)
			}

			switch message.Type {
			case ipv4.ICMPTypeEchoReply:
				handleEchoReply(message, host.(*net.IPAddr))
			case ipv4.ICMPTypeDestinationUnreachable:
				log.Println("Received destination unreachable")
			default:
				log.Println("Received unknown response")
			}
		}
	}()
}

func handleEchoReply(message *icmp.Message, host *net.IPAddr) {
	body := message.Body.(*icmp.Echo)
	log.Printf("Received response from %s with id %d and sequence %d", host, body.ID, body.Seq)

	// Search outstanding pings for a match
	for _, ping := range outstandingPings {
		if ping.PingerID == body.ID && ping.Sequence == body.Seq {
			handlePingMatch(ping)
			return
		}
	}
}

func handlePingMatch(ping Ping) {
	now := time.Now()
	ping.ReceivedAt = &now
	rtt := float64(ping.CalculateRoundTripTime().Milliseconds())

	rttGauge.With(
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
