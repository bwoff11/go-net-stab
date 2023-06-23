package main

import (
	config "github.com/bwoff11/go-net-stab/internal/config"
	metrics "github.com/bwoff11/go-net-stab/internal/metrics"
	log "github.com/sirupsen/logrus"
)

var pinger *Pinger

func main() {
	// Initialize and register metrics
	metrics := metrics.New()
	metrics.Register()
	log.Info("Initialized and registered metrics")

	// Load configuration
	config := config.New()
	if err := config.Load(); err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize the pinger and the listener
	pinger = &Pinger{
		Config:  config,
		Sent:    make(chan Ping, 100), // Buffered channel to prevent blocking
		Metrics: metrics,
	}

	listener := NewListener(pinger)
	listener.Start()
	log.Info("Initialized pinger and listener")

	// Establish ICMP connection
	// The createConnection() function from the pinger package creates a new ICMP connection that will be used to send Echo Requests and receive Echo Replies
	// If there is any error during this process, the program will terminate
	err := pinger.createConnection()
	if err != nil {
		log.Fatalf("Error creating ICMP connection: %v", err)
	}
	log.Info("Established ICMP connection")

	// Ensure the Timeout is correctly loaded and it's a positive value
	if pinger.Config.Timeout <= 0 {
		log.Fatal("Invalid Timeout in the configuration. It should be a positive value.")
	}

	// Start the pinging aprocesses
	go pinger.startPingingEndpoints()
	log.Info("Started pinging and listening processes")

	metrics.Expose("3009")
}
