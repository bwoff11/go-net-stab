package registry

import (
	"log"
	"net"
	"time"
)

type Ping struct {
	PingerID      int
	Sequence      int
	SourceIP      string
	DestinationIP string
	SentAt        time.Time
	ReceivedAt    *time.Time
}

var SentPings chan Ping
var outstandingPings []Ping

func Start() error {

	// Monitor the sentPings channel and import any new pings into the outstandingPings slice
	SentPings = make(chan Ping)
	go func() {
		for {
			ping := <-SentPings
			outstandingPings = append(outstandingPings, ping)
		}
	}()
	log.Println("Registry successfully started and waiting for pings")
	return nil
}

func HandleEchoReply(pingerID int, sequence int, host net.Addr) {
}
