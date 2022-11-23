package main

import (
	"log"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

type Config struct {
	Interval  int
	Endpoints []string
}

var config Config
var pingers []Pinger

func main() {
	createSentPingsChannel()
	registerPrometheusMetrics()
	readConfig()
	createPingers()

	for i := range pingers {
		go pingers[i].Run(config.Interval)
	}

	startResponseHandler()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":3009", nil)
}

func createPingers() {
	var id int
	for _, endpoint := range config.Endpoints {
		pingers = append(pingers, Pinger{
			ID:            id,
			SourceIP:      net.ParseIP("192.168.1.11"),
			DestinationIP: net.ParseIP(endpoint),
		})
		log.Println("Created new pinger for", endpoint, "with ID", id)
		id++
	}
}

func readConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}
}
