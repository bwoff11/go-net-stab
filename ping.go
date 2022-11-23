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

func (p *Ping) CalculateRoundTripTime() time.Duration {
	if p.ReceivedAt == nil {
		return 0
	}
	return p.ReceivedAt.Sub(p.SentAt)
}
