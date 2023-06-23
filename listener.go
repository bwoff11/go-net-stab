package main

import (
	"sync"
	"time"

	metrics "github.com/bwoff11/go-net-stab/internal/metrics"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Listener structure
type Listener struct {
	pinger   *Pinger
	received chan Ping
	pending  map[int]Ping
	mu       sync.Mutex
	Metrics  *metrics.Metrics
}

// NewListener initializes a new Listener.
func NewListener(p *Pinger) *Listener {
	return &Listener{
		pinger:   p,
		received: make(chan Ping, 256),
		pending:  make(map[int]Ping),
		Metrics:  p.Metrics, // assuming it's shared with the Pinger
	}
}

// Start function starts listening for sent and received pings.
func (l *Listener) Start() {
	log.Info("Started listening for pings")
	go l.handleSentPings()   // Handle sent pings
	go l.receivePings()      // Handle received pings
	go l.checkForLostPings() // check for lost pings
}

// handleSentPings function receives sent pings from the Sent channel and adds them to the pending map.
func (l *Listener) handleSentPings() {
	for sentPing := range l.pinger.Sent {
		l.mu.Lock()
		l.pending[sentPing.ID] = sentPing
		l.mu.Unlock()
	}
}

// receivePings function receives ICMP Echo Replies
func (l *Listener) receivePings() {
	packet := make([]byte, 1500)
	for {
		n, _, err := l.pinger.Connection.ReadFrom(packet)
		if err != nil {
			log.Error("Error reading from ICMP connection: ", err)
		}

		message, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), packet[:n])
		if err != nil {
			log.Error("Error parsing ICMP message: ", err)
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

			l.Metrics.ReceivedPingsCounter.WithLabelValues(
				l.pinger.Config.Localhost,
				l.pinger.Config.Endpoints[receivedPing.ID].Hostname,
				l.pinger.Config.Endpoints[receivedPing.ID].Address,
				l.pinger.Config.Endpoints[receivedPing.ID].Location,
			).Inc()

			l.Metrics.RttGauge.WithLabelValues(
				l.pinger.Config.Localhost,
				l.pinger.Config.Endpoints[receivedPing.ID].Hostname,
				l.pinger.Config.Endpoints[receivedPing.ID].Address,
				l.pinger.Config.Endpoints[receivedPing.ID].Location,
			).Set(float64(time.Since(sentPing.SentAt).Nanoseconds()) / 1000000)

			if ok {
				receivedPing.SentAt = sentPing.SentAt
				l.received <- receivedPing
				log.WithFields(log.Fields{
					"ID":     receivedPing.ID,
					"Seq":    receivedPing.Seq,
					"SentAt": receivedPing.SentAt,
				}).Info("Received ping")
			} else {
				log.WithFields(log.Fields{
					"ID":  receivedPing.ID,
					"Seq": receivedPing.Seq,
				}).Error("Received ping but no corresponding request found")
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

			log.WithFields(log.Fields{
				"ID":     ping.ID,
				"Seq":    ping.Seq,
				"SentAt": ping.SentAt,
			}).Error("Lost ping")

			delete(l.pending, id)
			log.WithFields(log.Fields{
				"ID":  id,
				"Seq": ping.Seq,
			}).Info("Removed pending ping")
		}
	}
}
