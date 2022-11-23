package main

import (
	"time"

	"github.com/bwoff11/go-net-stab/internal/config"
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

func (p *Ping) IsTimedOut() bool {
	now := time.Now()
	threshold := config.Config.Timeout
	sinceSent := now.Sub(p.SentAt).Seconds()
	return sinceSent > float64(threshold)
}
