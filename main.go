package main

import (
	metrics "github.com/bwoff11/go-net-stab/internal/metrics"
	log "github.com/sirupsen/logrus"
)

var pinger *Pinger
var listener *Listener

func main() {
	metrics := metrics.New()
	metrics.Register()
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

	// Load the configuration
	if err := pinger.loadConfig(); err != nil {
		log.Fatal(err)
	}

	// Ensure the Timeout is correctly loaded and it's a positive value
	if pinger.Config.Timeout <= 0 {
		log.Fatal("Invalid Timeout in the configuration. It should be a positive value.")
	}

	// Start the pinging aprocesses
	go pinger.startPingingEndpoints()
	log.Info("Started pinging and listening processes")

	metrics.Expose("3009")
}
