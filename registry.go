package main

import (
	"log"
	"time"
)

var sentPings chan Ping
var outstandingPings []Ping

func createSentPingsChannel() {
	sentPings = make(chan Ping)
	go func() {
		for {
			ping := <-sentPings
			outstandingPings = append(outstandingPings, ping)
		}
	}()
}

func startLostPacketChecker() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			for _, ping := range outstandingPings {
				if ping.IsTimedOut() {
					log.Println("Ping expired")
				}
			}
		}
	}()
}
