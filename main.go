package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Configuration holds all the configuration values
type Configuration struct {
	Interval  time.Duration // Time interval in milliseconds between each ping
	Timeout   time.Duration // Time in milliseconds to wait for a ping response before considering it lost
	Port      string        // Port for the HTTP server that exposes the metrics
	Localhost string        // Hostname of the machine where this program is running
	Endpoints []Endpoint    // List of endpoints to ping
}

// Endpoint represents each endpoint that will be pinged
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

// Global variables used across multiple functions
var (
	config *Configuration   // holds the current configuration
	conn   *icmp.PacketConn // ICMP connection used to send and receive pings
	sent   chan Ping        // channel to keep track of sent pings
	pending []Ping          // slice to keep track of pings that have been sent but not yet received
	mutex  sync.Mutex       // mutex for synchronizing access to pending slice
)

// Ping represents a single ping
type Ping struct {
	ID     int
	Seq    int
	SentAt time.Time
}

var (
	SentPingsCounter = prometheus.NewCounterVec(
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

	LostPingsCounter = prometheus.NewCounterVec(
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

	RttGauge = prometheus.NewGaugeVec(
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
)

func main() {
	config = &Configuration{}
	if err := LoadConfig(); err != nil {
		log.Fatal(err)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatal(err)
	}

	// Start a goroutine for each endpoint to continuously send pings
	var wg sync.WaitGroup
	for i, endpoint := range config.Endpoints {
		wg.Add(1)
		go func(id int, endpoint Endpoint) {
			defer wg.Done()
			ping(id, endpoint)
		}(i, endpoint)
	}

	sent = make(chan Ping)
	go func() {
		for {
			pending = append(pending, <-sent)
		}
	}()

	createListener()
	checkForLostPings()

	// Register the Prometheus metrics and start the HTTP server to serve the metrics
	prometheus.MustRegister(SentPingsCounter)
	prometheus.MustRegister(LostPingsCounter)
	prometheus.MustRegister(RttGauge)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))

	wg.Wait()
}

// LoadConfig loads the configuration from a file
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

// ping continuously sends pings to an endpoint
func ping(id int, endpoint Endpoint) {
	sequence := 0
	ticker := time.NewTicker(config.Interval)

	for range ticker.C {
		m := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   id,
				Seq:  sequence,
				Data: []byte("we've been trying to reach you about your car's extended warranty"),
			},
			Src: net.IPv4zero,
		}

		b, err := m.Marshal(nil)
		if err != nil {
			log.Fatal(err)
		}

		_, err = conn.WriteTo(b, &net.IPAddr{IP: net.ParseIP(endpoint.Address)})
		if err != nil {
			log.Println("Error sending ping to", endpoint.Hostname, "at", endpoint.Address, "at", endpoint.Location, ":", err)
			break
		}

		sent <- Ping{
			ID:     id,
			Seq:    sequence,
			SentAt: time.Now(),
		}

		SentPingsCounter.With(
			prometheus.Labels{
				"source_hostname":      config.Localhost,
				"destination_hostname": endpoint.Hostname,
				"destination_address":  endpoint.Address,
				"destination_location": endpoint.Location,
			},
		).Inc()

		sequence++

		log.Println("Sent ping to", endpoint.Hostname, "at", endpoint.Address, "at", endpoint.Location)
	}
}

// createListener listens for incoming ICMP echo replies
func createListener() {
	go func() {
		rb := make([]byte, 1500)

		for {
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
				mutex.Lock()
				for i, ping := range pending {
					if ping.ID == rm.Body.(*icmp.Echo).ID && ping.Seq == rm.Body.(*icmp.Echo).Seq {
						pending = append(pending[:i], pending[i+1:]...)
						rtt := time.Since(ping.SentAt).Milliseconds()

						RttGauge.With(
							prometheus.Labels{
								"source_hostname":      config.Localhost,
								"destination_hostname": config.Endpoints[rm.Body.(*icmp.Echo).ID].Hostname,
								"destination_address":  config.Endpoints[rm.Body.(*icmp.Echo).ID].Address,
								"destination_location": config.Endpoints[rm.Body.(*icmp.Echo).ID].Location,
							},
						).Set(float64(rtt))

						log.Println("Received ping reply from", config.Endpoints[rm.Body.(*icmp.Echo).ID].Hostname, "at", config.Endpoints[rm.Body.(*icmp.Echo).ID].Address, "at", config.Endpoints[rm.Body.(*icmp.Echo).ID].Location, "with RTT", rtt, "ms")
						break
					}
				}
				mutex.Unlock()
			}
		}
	}()
}

// checkForLostPings checks the pending slice for lost pings
func checkForLostPings() {
	ticker := time.NewTicker(50 * time.Millisecond)

	go func() {
		for range ticker.C {
			mutex.Lock()
			for i := 0; i < len(pending); {
				ping := pending[i]
				if time.Since(ping.SentAt) > config.Timeout {
					log.Println("Ping to", config.Endpoints[ping.ID].Hostname, "at", config.Endpoints[ping.ID].Address, "in", config.Endpoints[ping.ID].Location, "timed out")

					LostPingsCounter.With(
						prometheus.Labels{
							"source_hostname":      config.Localhost,
							"destination_hostname": config.Endpoints[ping.ID].Hostname,
							"destination_address":  config.Endpoints[ping.ID].Address,
							"destination_location": config.Endpoints[ping.ID].Location,
						},
					).Inc()

					pending[i] = pending[len(pending)-1]
					pending = pending[:len(pending)-1]
				} else {
					i++
				}
			}
			mutex.Unlock()
		}
	}()
}
