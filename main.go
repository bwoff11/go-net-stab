package main

import (
	"log"
	"net/http"

	"github.com/bwoff11/go-net-stab/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var pingers []Pinger

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatal("Failed to load config:", err)
	}
	createSentPingsChannel()
	registerPrometheusMetrics()
	createPingers()
	startResponseHandler()

	for i := range pingers {
		go pingers[i].Run(config.Config.Interval)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":3009", nil)
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
