package main

import (
	"log"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Listener structure
type Listener struct {
	pinger   *Pinger
	received chan Ping
	pending  map[int]Ping
	mu       sync.Mutex
	Metrics  *Metrics
}

// listenForPings function listens for ICMP Echo Replies and matches them to their corresponding
// ICMP Echo Requests. If a Reply is not received for a particular Request within the
// specified timeout, that Request is considered as lost.
func (l *Listener) listenForPings() {
	go l.receivePings() // calling receivePings in a separate goroutine

	ticker := time.NewTicker(l.pinger.Config.Timeout)

	for range ticker.C {
		l.checkForLostPings()
	}
}

// receivePings function receives ICMP Echo Replies
func (l *Listener) receivePings() {
	packet := make([]byte, 1500)
	for {
		n, _, err := pinger.Connection.ReadFrom(packet)
		if err != nil {
			log.Fatalf("Error receiving ICMP packet: %v", err)
		}

		message, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), packet[:n])
		if err != nil {
			log.Printf("Error parsing ICMP message: %v", err)
			continue
		}

		switch message.Type {
		case ipv4.ICMPTypeEchoReply:
			body := message.Body.(*icmp.Echo)
			receivedPing := Ping{
				ID:     body.ID,
				Seq:    body.Seq,
				SentAt: time.Now(), // This will be overwritten with the correct value
			}

			l.mu.Lock()
			sentPing, ok := l.pending[receivedPing.ID]
			l.mu.Unlock()

			if ok {
				receivedPing.SentAt = sentPing.SentAt
				l.received <- receivedPing
			}
		}
	}
}

// checkForLostPings function checks for any lost pings
func (l *Listener) checkForLostPings() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for id, ping := range l.pending {
		if time.Since(ping.SentAt) > l.pinger.Config.Timeout {
			l.Metrics.LostPingsCounter.WithLabelValues(
				l.pinger.Config.Localhost,
				l.pinger.Config.Endpoints[ping.ID].Hostname,
				l.pinger.Config.Endpoints[ping.ID].Address,
				l.pinger.Config.Endpoints[ping.ID].Location,
			).Inc()

			delete(l.pending, id)
		}
	}
}
