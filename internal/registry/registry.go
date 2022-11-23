package registry

import (
	"log"
	"time"
)

type Registry struct {
	PingsSent       chan Ping
	PingsPending    []Ping
	PingsHistorical []Ping
}

var SentPings chan Ping
var pending []Ping

func Start() error {

	startPendingImporter()
	startLostPingChecker()

	log.Println("Registry successfully started and is now waiting for pings from pingers")
	return nil
}

func startPendingImporter() {
	SentPings = make(chan Ping)
	go func() {
		for {
			ping := <-SentPings
			pending = append(pending, ping)
		}
	}()
}

func startLostPingChecker() {
	go func() {
		for {
			for _, ping := range pending {
				if ping.IsLost() {
					handleLostPing(ping)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
}
