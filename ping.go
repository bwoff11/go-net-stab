package main

import (
	"net"
	"time"
)

type Ping struct {
	PingerID      int
	SourceIP      net.IP
	DestinationIP net.IP
	Sequence      int
	SentAt        time.Time
	ReceivedAt    *time.Time
}

func NewPing(pingerID int, sourceIP net.IP, destinationIP net.IP, sequence int) Ping {
	return Ping{
		PingerID:      pingerID,
		SourceIP:      sourceIP,
		DestinationIP: destinationIP,
		Sequence:      sequence,
		SentAt:        time.Now(),
		ReceivedAt:    nil,
	}
}

func (p *Ping) CalculateRoundTripTime() time.Duration {
	if p.ReceivedAt == nil {
		return 0
	}
	return p.ReceivedAt.Sub(p.SentAt)
}
