package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var pinger *Pinger
var listener *Listener

func main() {
	// Initialize and register metrics
	// NewMetrics() is a function from the metrics package that initializes a new Metrics struct
	// The RegisterMetrics() function from the metrics package then registers these metrics with Prometheus
	metrics := NewMetrics()
	metrics.RegisterMetrics()
	log.Info("Initialized and registered metrics")

	// Initialize the pinger and the listener
	// The pinger is responsible for sending ICMP Echo Requests to the specified endpoints
	// The listener is responsible for receiving ICMP Echo Replies and matching them to their corresponding Requests
	// Both pinger and listener have buffered channels to prevent blocking. The buffer size of 100 is arbitrary and can be adjusted as needed
	pinger = &Pinger{
		Config:  &Configuration{},
		Sent:    make(chan Ping, 100), // Buffered channel to prevent blocking
		Metrics: metrics,
	}

	listener = &Listener{
		pinger:   pinger,
		received: make(chan Ping, 100), // Buffered channel to prevent blocking
		pending:  make(map[int]Ping),
		Metrics:  metrics,
	}
	log.Info("Initialized pinger and listener")

	// Establish ICMP connection
	// The createConnection() function from the pinger package creates a new ICMP connection that will be used to send Echo Requests and receive Echo Replies
	// If there is any error during this process, the program will terminate
	err := pinger.createConnection()
	if err != nil {
		log.Fatalf("Error creating ICMP connection: %v", err)
	}
	log.Info("Established ICMP connection")

	// Load the configuration
	if err := pinger.loadConfig(); err != nil {
		log.Fatal(err)
	}

	// Ensure the Timeout is correctly loaded and it's a positive value
	if pinger.Config.Timeout <= 0 {
		log.Fatal("Invalid Timeout in the configuration. It should be a positive value.")
	}

	// Start the pinging and listening processes
	// The startPingingEndpoints() function from the pinger package starts the process of sending ICMP Echo Requests to the specified endpoints
	// The listenForPings() function from the listener package starts the process of receiving ICMP Echo Replies and matching them to their corresponding Requests
	// Both functions run in separate goroutines to enable concurrent execution
	go pinger.startPingingEndpoints()
	go listener.listenForPings()

	log.Info("Started pinging and listening processes")

	// Expose the metrics endpoint
	// The /metrics endpoint is where Prometheus will scrape the metrics data from
	// The promhttp.Handler() function from the Prometheus client library provides a HTTP handler to expose the registered metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":3009", nil))
}
