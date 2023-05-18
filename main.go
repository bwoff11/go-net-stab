package main

import (
	"errors"
	"log"
	"net"
	"net/http" // to expose a HTTP server for Prometheus to scrape metrics from
	"time"

	"github.com/prometheus/client_golang/prometheus" // for creating Prometheus metrics
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"golang.org/x/net/icmp" // ICMP ping operations
	"golang.org/x/net/ipv4" // ICMP ping operations
)

// Configuration struct to hold all configuration values
type Configuration struct {
	Interval  int         // Time interval in milliseconds between each ping
	Timeout   int         // Time in milliseconds to wait for a ping response before considering it lost
	Port      string      // Port for the HTTP server that exposes the metrics
	Localhost string      // Hostname of the machine where this program is running
	Endpoints []Endpoint  // List of endpoints to ping
}

// Endpoint struct to represent each endpoint that will be pinged
type Endpoint struct {
	Hostname string
	Address  string
	Location string
}

// List of configuration file locations that will be checked in order
var configLocations = []string{
	"/etc/go-net-stab/",
	"$HOME/.config/go-net-stab/",
	"$HOME/.go-net-stab",
	".",
}

// Prometheus counters for tracking sent and lost pings
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

// Prometheus gauge for tracking round trip time (RTT)
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

// Global variables used across multiple functions
var config *Configuration  // holds the current configuration
var conn *icmp.PacketConn  // ICMP connection used to send and receive pings
var sent chan Ping         // channel to keep track of sent pings
var pending []Ping         // slice to keep track of pings that have been sent but not yet received

// Struct representing a single ping
type Ping struct {
	ID     int
	Seq    int
	SentAt time.Time
}

func main() {
	// Load the configuration
	config = &Configuration{}
	if err := LoadConfig(); err != nil {
		log.Fatal(err)
	}

	// Listen for ICMP packets
	var err error
	conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatal(err)
	}

	// Start a goroutine for each endpoint to continuously send pings
	for i, endpoint := range config.Endpoints {
		go ping(i, endpoint)
	}

	// Initialize the channel for sent pings and start a goroutine to add each sent ping to the pending slice
	sent = make(chan Ping)
	go func() {
		for {
			pending = append(pending, <-sent)
		}
	}()

	// Start a goroutine to listen for incoming ICMP echo replies
	createListener()

	// Start a goroutine to check the pending slice for lost pings
	checkForLostPings()

	// Register the Prometheus metrics and start the HTTP server to serve the metrics
	prometheus.MustRegister(SentPingsCounter)
	prometheus.MustRegister(LostPingsCounter)
	prometheus.MustRegister(RttGauge)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":3009", nil))
}

// Function to load the configuration from a file
func LoadConfig() error {
	// Set the name of the config file
	viper.SetConfigName("config")

	// Add all the possible locations of the config file
	for _, location := range configLocations {
		viper.AddConfigPath(location)
	}

	// Try to read the config file
	if err := viper.ReadInConfig(); err != nil {
		return errors.New("Fatal error config file: " + err.Error())
	}

	// Try to unmarshal the config file into the Configuration struct
	if err := viper.Unmarshal(config); err != nil {
		return errors.New("Fatal error config file: " + err.Error())
	}

	// Log all the loaded endpoints
	for _, endpoint := range config.Endpoints {
		log.Println("Loaded endpoint", endpoint.Hostname, "at", endpoint.Address, "at", endpoint.Location)
	}

	log.Println("Configuration successfully loaded from", viper.ConfigFileUsed())
	return nil
}

// Function to continuously send pings to an endpoint
func ping(id int, endpoint Endpoint) {
	// Initialize variables
	sequence := 0
	tickerTime := time.Duration(config.Interval) * time.Millisecond
	ticker := time.NewTicker(tickerTime)

	// Start a goroutine to send a ping every time the ticker ticks
	go func() {
		for {
			<-ticker.C

			// Create the ICMP echo request message
			m := icmp.Message{
				Type: ipv4.ICMPTypeEcho, Code: 0,
				Body: &icmp.Echo{
					ID:   id,
					Seq:  sequence,
					Data: []byte("we've been trying to reach you about your car's extended warranty"),
				},
			}

			// Marshal the message into a byte slice
			b, err := m.Marshal(nil)
			if err != nil {
				log.Fatal(err)
			}

			// Send the ICMP echo request
			_, err = conn.WriteTo(b, &net.IPAddr{IP: net.ParseIP(endpoint.Address)})
			if err != nil {
				log.Println("Error sending ping to", endpoint.Hostname, "at", endpoint.Address, "at", endpoint.Location, ":", err)
				break
			}

			// Add the ping to the sent channel
			sent <- Ping{
				ID:     id,
				Seq:    sequence,
				SentAt: time.Now(),
			}

			// Increment the sent pings counter
			SentPingsCounter.With(
				prometheus.Labels{
					"source_hostname":      config.Localhost,
					"destination_hostname": endpoint.Hostname,
					"destination_address":  endpoint.Address,
					"destination_location": endpoint.Location,
				},
			).Inc()

			// Increment the sequence number for the next ping
			sequence++

			// Log that a ping was sent
			log.Println("Sent ping to", endpoint.Hostname, "at", endpoint.Address, "at", endpoint.Location)
		}
	}()
}

// Function to listen for incoming ICMP echo replies
func createListener() {
	// Start a goroutine to continuously read incoming packets
	go func() {
		for {
			// Initialize the byte slice to read the packet into
			rb := make([]byte, 1500)

			// Read the packet
			n, _, err := conn.ReadFrom(rb)
			if err != nil {
				log.Fatal(err)
			}

			// Parse the ICMP message from the packet
			rm, err := icmp.ParseMessage(1, rb[:n])
			if err != nil {
				log.Fatal(err)
			}

			// Check if the message is an ICMP echo reply
			switch rm.Type {
			case ipv4.ICMPTypeEchoReply:
				// Look for the corresponding ping in the pending slice
				for i, ping := range pending {
					if ping.ID == rm.Body.(*icmp.Echo).ID && ping.Seq == rm.Body.(*icmp.Echo).Seq {
						// Remove the corresponding ping from the pending slice
						pending = append(pending[:i], pending[i+1:]...)

						// Calculate the round-trip time
						rtt := time.Since(ping.SentAt).Milliseconds()

						// Set the RTT gauge to the calculated RTT
						RttGauge.With(
							prometheus.Labels{
								"source_hostname":      config.Localhost,
								"destination_hostname": config.Endpoints[rm.Body.(*icmp.Echo).ID].Hostname,
								"destination_address":  config.Endpoints[rm.Body.(*icmp.Echo).ID].Address,
								"destination_location": config.Endpoints[rm.Body.(*icmp.Echo).ID].Location,
							},
						).Set(float64(rtt))

						// Log that an ICMP echo reply was received
						log.Println("Received ping reply from", config.Endpoints[rm.Body.(*icmp.Echo).ID].Hostname, "at", config.Endpoints[rm.Body.(*icmp.Echo).ID].Address, "at", config.Endpoints[rm.Body.(*icmp.Echo).ID].Location, "with RTT", rtt, "ms")
					}
				}
			}
		}
	}()
}

// Function to check the pending slice for lost pings
func checkForLostPings() {
	// Initialize the ticker
	ticker := time.NewTicker(50 * time.Millisecond)

	// Start a goroutine to check the pending slice every time the ticker ticks
	go func() {
		for {
			<-ticker.C
			for i, ping := range pending {
				// Check if the ping is older than the timeout value
				if time.Since(ping.SentAt) > time.Duration(config.Timeout)*time.Millisecond {
					// Log that a ping was lost
					log.Println("Ping to", config.Endpoints[ping.ID].Hostname, "at", config.Endpoints[ping.ID].Address, "in", config.Endpoints[ping.ID].Location, "timed out")

					// Increment the lost pings counter
					LostPingsCounter.With(
						prometheus.Labels{
							"source_hostname":      config.Localhost,
							"destination_hostname": config.Endpoints[ping.ID].Hostname,
							"destination_address":  config.Endpoints[ping.ID].Address,
							"destination_location": config.Endpoints[ping.ID].Location,
						},
					).Inc()

					// Remove the ping from the pending slice
					pending = append(pending[:i], pending[i+1:]...)
					
					// Break here since removing from the slice will mess up the loop and cause a crash
					// Interval should be fast enough even if all pings are lost so this should be fine for now
					break
				}
			}
		}
	}()
}
