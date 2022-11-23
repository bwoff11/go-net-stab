package pingers

import (
	"log"

	"github.com/bwoff11/go-net-stab/internal/config"
)

var pingers []Pinger

func Start() error {
	createPingers()
	startPingers()
	return nil
}

func createPingers() {
	endpoints := config.Config.Endpoints
	log.Println("Creating pingers for", len(endpoints), "endpoints")

	var id int
	for _, endpoint := range endpoints {
		newPinger := Pinger{
			ID:            id,
			SourceIP:      "192.168.1.11",
			DestinationIP: endpoint,
			Sequence:      0,
		}
		pingers = append(pingers, newPinger)
		log.Println("Created pinger", newPinger.ID, "for", newPinger.DestinationIP)
		id++
	}
}

func startPingers() {
	for i := range pingers {
		go pingers[i].Start()
	}
}
