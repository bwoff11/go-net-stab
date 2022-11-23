package registry

import "net"

var sentPings chan Ping
var outstandingPings []Ping

func Start() error {

	// Monitor the sentPings channel and import any new pings into the outstandingPings slice
	sentPings = make(chan Ping)
	go func() {
		for {
			ping := <-sentPings
			outstandingPings = append(outstandingPings, ping)
		}
	}()
}

func HandleEchoReply(pingerID int, sequence int, host net.Addr) {
}
