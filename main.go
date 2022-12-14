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
	[]string{
		"source_hostname",
		"destination_hostname",
		"destination_address",
		"destination_location",
	},
)

var LostPingsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ping_lost_packet_total",
		Help: "Total number of packets lost",
	},
	[]string{
		"source_hostname",
		"destination_hostname",
		"destination_address",
		"destination_location",
	},
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
	tickerTime := time.Duration(config.Interval) * time.Millisecond
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
				log.Println("Error sending ping to", endpoint.Hostname, "at", endpoint.Address, "at", endpoint.Location, ":", err)
				break
			}

			// Add to sent
			sent <- Ping{
				ID:     id,
				Seq:    sequence,
				SentAt: time.Now(),
			}

			// Update metrics
			SentPingsCounter.With(
				prometheus.Labels{
					"source_hostname":      config.Localhost,
					"destination_hostname": endpoint.Hostname,
					"destination_address":  endpoint.Address,
					"destination_location": endpoint.Location,
				},
			).Inc()

			// Increment sequence
			sequence++

			// Log
			log.Println("Sent ping to", endpoint.Hostname, "at", endpoint.Address, "at", endpoint.Location)
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
						RttGauge.With(
							prometheus.Labels{
								"source_hostname":      config.Localhost,
								"destination_hostname": config.Endpoints[rm.Body.(*icmp.Echo).ID].Hostname,
								"destination_address":  config.Endpoints[rm.Body.(*icmp.Echo).ID].Address,
								"destination_location": config.Endpoints[rm.Body.(*icmp.Echo).ID].Location,
							},
						).Set(float64(rtt))

						// Log
						log.Println("Received ping reply from", config.Endpoints[rm.Body.(*icmp.Echo).ID].Hostname, "at", config.Endpoints[rm.Body.(*icmp.Echo).ID].Address, "at", config.Endpoints[rm.Body.(*icmp.Echo).ID].Location, "with RTT", rtt, "ms")
					}
				}
			}
		}
	}()
}

// Check the pending queue for any ping that is exceeding the timeout value.
// If so, increment the lost pings counter and remove from pending.
func checkForLostPings() {
	ticker := time.NewTicker(50 * time.Millisecond)
	go func() {
		for {
			<-ticker.C
			for i, ping := range pending {
				if time.Since(ping.SentAt) > time.Duration(config.Timeout)*time.Millisecond {
					log.Println("Ping to", config.Endpoints[ping.ID].Hostname, "at", config.Endpoints[ping.ID].Address, "in", config.Endpoints[ping.ID].Location, "timed out")

					// Increment lost pings counter
					LostPingsCounter.With(
						prometheus.Labels{
							"source_hostname":      config.Localhost,
							"destination_hostname": config.Endpoints[ping.ID].Hostname,
							"destination_address":  config.Endpoints[ping.ID].Address,
							"destination_location": config.Endpoints[ping.ID].Location,
						},
					).Inc()

					// Remove from pending
					pending = append(pending[:i], pending[i+1:]...)

					// Break here since removing from the slice will mess up the loop and cause a crash
					// Interval should be fast enough even if all pings are lost so this should be fine for now
					break
				}
			}
		}
	}()
}
