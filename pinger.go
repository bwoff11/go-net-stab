package main

import (
	"net"
	"sync"
	"time"

	config "github.com/bwoff11/go-net-stab/internal/config"
	metrics "github.com/bwoff11/go-net-stab/internal/metrics"
	log "github.com/sirupsen/logrus"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Pinger structure
type Pinger struct {
	Config     *config.Configuration
	Connection *icmp.PacketConn
	Sent       chan Ping
	Metrics    *metrics.Metrics
}

// Ping structure
type Ping struct {
	ID     int
	Seq    int
	SentAt time.Time
}

func New(config *config.Configuration, metrics *metrics.Metrics) *Pinger {
	return &Pinger{
		Config:  config,
		Sent:    make(chan Ping, 256),
		Metrics: metrics,
	}
}

func (p *Pinger) createConnection() error {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}

	p.Connection = conn
	return nil
}

func (p *Pinger) startPingingEndpoints() {
	var wg sync.WaitGroup
	for i, endpoint := range p.Config.Endpoints {
		wg.Add(1)
		go func(id int, endpoint config.Endpoint) {
			defer wg.Done()
			p.pingEndpoint(id, endpoint)
		}(i, endpoint)
	}
	wg.Wait() // Add this to wait for all goroutines to finish
}

func (p *Pinger) pingEndpoint(id int, endpoint config.Endpoint) {
	sequence := 0
	ticker := time.NewTicker(time.Millisecond * p.Config.Interval)

	for range ticker.C {
		m := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   id, // Use `id` for Echo ID
				Seq:  sequence,
				Data: []byte("we've been trying to reach you about your car's extended warranty"),
			},
		}

		b, err := m.Marshal(nil)
		if err != nil {
			log.Fatal(err)
		}

		_, err = p.Connection.WriteTo(b, &net.IPAddr{IP: net.ParseIP(endpoint.Address)})
		if err != nil {
			log.Printf("Error sending ping to %s at %s in %s: %v", endpoint.Hostname, endpoint.Address, endpoint.Location, err)
			break
		}

		sentPing := Ping{
			ID:     m.Body.(*icmp.Echo).ID, // Use Echo ID as Ping ID
			Seq:    sequence,
			SentAt: time.Now(),
		}
		p.Sent <- sentPing

		log.WithFields(log.Fields{
			"ID":  sentPing.ID,
			"Seq": sequence,
		}).Info("Sent ping")

		p.Metrics.SentPingsCounter.WithLabelValues(
			p.Config.Localhost,
			endpoint.Hostname,
			endpoint.Address,
			endpoint.Location,
		).Inc()

		sequence++
	}
}
