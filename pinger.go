package main

import (
	"io/ioutil"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"gopkg.in/yaml.v2"
)

// Configuration structure
type Configuration struct {
	Interval  time.Duration `yaml:"interval"`
	Timeout   time.Duration `yaml:"timeout"`
	Port      string        `yaml:"port"`
	Localhost string        `yaml:"localhost"`
	Endpoints []Endpoint    `yaml:"endpoints"`
}

// Endpoint structure
type Endpoint struct {
	Hostname string        `yaml:"hostname"`
	Address  string        `yaml:"address"`
	Location string        `yaml:"location"`
	Interval time.Duration `yaml:"interval"`
}

// Pinger structure
type Pinger struct {
	Config     *Configuration
	Connection *icmp.PacketConn
	Sent       chan Ping
	Metrics    *Metrics
}

// Ping structure
type Ping struct {
	ID     int
	Seq    int
	SentAt time.Time
}

func (p *Pinger) loadConfig() error {
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(configFile, p.Config)
	if err != nil {
		return err
	}

	// Convert from milliseconds to nanoseconds
	p.Config.Interval *= time.Millisecond
	p.Config.Timeout *= time.Millisecond

	return nil
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
		go func(id int, endpoint Endpoint) {
			defer wg.Done()
			p.pingEndpoint(id, endpoint)
		}(i, endpoint)
	}
	wg.Wait() // Add this to wait for all goroutines to finish
}

func (p *Pinger) pingEndpoint(id int, endpoint Endpoint) {
	sequence := 0
	ticker := time.NewTicker(p.Config.Interval)

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
