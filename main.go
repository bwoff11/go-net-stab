package main

import (
	"log"

	"github.com/bwoff11/go-net-stab/internal/config"
	"github.com/bwoff11/go-net-stab/internal/listener"
	"github.com/bwoff11/go-net-stab/internal/reporting"
)

var pingers []Pinger

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatal("Failed to load config:", err)
	}
	if err := listener.Start(); err != nil {
		log.Fatal("Failed to start listener:", err)
	}
	createSentPingsChannel()
	createPingers()

	for i := range pingers {
		go pingers[i].Run(config.Config.Interval)
	}

	reporting.ServeMetrics()
}

func createPingers() {
	var id int
	for _, endpoint := range config.Config.Endpoints {
		pingers = append(pingers, Pinger{
			ID:            id,
			SourceIP:      "192.168.1.11",
			DestinationIP: endpoint,
		})
		log.Println("Created new pinger for", endpoint, "with ID", id)
		id++
	}
}
