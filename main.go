package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Configuration struct {
	Interval  int
	Timeout   int
	Port      string
	Localhost string
	Endpoints []Endpoint
}

type Endpoint struct {
	Hostname string
	Address  string
	Location string
}

var configLocations = []string{
	"/etc/go-net-stab/",
	"$HOME/.config/go-net-stab/",
	"$HOME/.go-net-stab",
	".",
}

var SentPingsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ping_sent_packet_total",
		Help: "Total number of packets sent",
	},
	[]string{"source_ip", "destination_ip"},
)

var LostPingsCounter = prometheus.NewCounterVec(
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
	[]string{
		"source_hostname",
		"destination_hostname",
		"destination_address",
		"destination_location",
	},
)

var config *Configuration
var conn *icmp.PacketConn
var sent chan Ping
var pending []Ping

type Ping struct {
	ID     int
	Seq    int
	SentAt time.Time
}

func main() {
	config = &Configuration{}
	if err := LoadConfig(); err != nil {
		log.Fatal(err)
	}

	var err error
	conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatal(err)
	}

	for i, endpoint := range config.Endpoints {
		go ping(i, endpoint)
	}

	// Init sent channel and prepetually move to pending
	sent = make(chan Ping)
	go func() {
		for {
			pending = append(pending, <-sent)
		}
	}()

	createListener()
	checkForLostPings()

	// ServeMetrics
	prometheus.MustRegister(SentPingsCounter)
	prometheus.MustRegister(LostPingsCounter)
	prometheus.MustRegister(RttGauge)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":3009", nil))
}

func LoadConfig() error {
	viper.SetConfigName("config")

	for _, location := range configLocations {
		viper.AddConfigPath(location)
	}
	if err := viper.ReadInConfig(); err != nil {
		return errors.New("Fatal error config file: " + err.Error())
	}
	if err := viper.Unmarshal(config); err != nil {
		return errors.New("Fatal error config file: " + err.Error())
	}

	for _, endpoint := range config.Endpoints {
		log.Println("Loaded endpoint", endpoint.Hostname, "at", endpoint.Address, "at", endpoint.Location)
	}

	log.Println("Configuration successfully loaded from", viper.ConfigFileUsed())
	return nil
}

func ping(id int, endpoint Endpoint) {

	sequence := 0
	tickerTime := time.Duration(config.Interval) * time.Second
	ticker := time.NewTicker(tickerTime)

	go func() {
		for {
			<-ticker.C

			// Create ICMP message
			m := icmp.Message{
				Type: ipv4.ICMPTypeEcho, Code: 0,
				Body: &icmp.Echo{
					ID:   id,
					Seq:  sequence,
					Data: []byte("we've been trying to reach you about your car's extended warranty"),
				},
			}

			// Marshal message
			b, err := m.Marshal(nil)
			if err != nil {
				log.Fatal(err)
			}

			// Send message
			_, err = conn.WriteTo(b, &net.IPAddr{IP: net.ParseIP(endpoint.Address)})
			if err != nil {
				log.Fatal(err)
			}

			// Add to sent
			sent <- Ping{
				ID:     id,
				Seq:    sequence,
				SentAt: time.Now(),
			}

			// Increment sequence
			sequence++
		}
	}()
}

func createListener() {
	go func() {
		for {
			rb := make([]byte, 1500)
			n, _, err := conn.ReadFrom(rb)
			if err != nil {
				log.Fatal(err)
			}
			rm, err := icmp.ParseMessage(1, rb[:n])
			if err != nil {
				log.Fatal(err)
			}
			switch rm.Type {
			case ipv4.ICMPTypeEchoReply:

				// Find matching pending ping
				for i, ping := range pending {
					if ping.ID == rm.Body.(*icmp.Echo).ID && ping.Seq == rm.Body.(*icmp.Echo).Seq {
						// Remove from pending
						pending = append(pending[:i], pending[i+1:]...)

						// Calculate RTT
						rtt := time.Since(ping.SentAt).Milliseconds()

						// Update metrics
						SentPingsCounter.WithLabelValues(config.Localhost, config.Endpoints[ping.ID].Address).Inc()
						RttGauge.WithLabelValues(
							config.Localhost,
							config.Endpoints[ping.ID].Hostname,
							config.Endpoints[ping.ID].Address,
							config.Endpoints[ping.ID].Location,
						).Set(float64(rtt))

						log.Println("Received ping reply:", ping.ID, ping.Seq, "RTT:", rtt)
					}
				}
			}
		}
	}()
}

func checkForLostPings() {

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			<-ticker.C
			for _, ping := range pending {
				if time.Since(ping.SentAt) > time.Duration(config.Timeout)*time.Second {
					log.Println("Ping to", config.Endpoints[ping.ID].Hostname, "at", config.Endpoints[ping.ID].Address, "in", config.Endpoints[ping.ID].Location, "timed out")
				}
			}
		}
	}()
}
