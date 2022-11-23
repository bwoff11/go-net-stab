package main

import (
	"time"
)

type Ping struct {
	PingerID      int
	SourceIP      string
	DestinationIP string
	Sequence      int
	SentAt        time.Time
	ReceivedAt    *time.Time
}

func NewPing(pingerID int, sourceIP string, destinationIP string, sequence int) Ping {
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
