package registry

import (
	"time"

	"github.com/bwoff11/go-net-stab/internal/config"
)

type Ping struct {
	PingerID      int
	Sequence      int
	SourceIP      string
	DestinationIP string
	SentAt        time.Time
	ReceivedAt    *time.Time
}

func (p *Ping) CalculateRoundTripTime() time.Duration {
	if p.ReceivedAt == nil {
		return 0
	}
	return p.ReceivedAt.Sub(p.SentAt)
}

// Checks to see if the time since the ping was sent exceeds the configured timeout
func (p *Ping) IsLost() bool {
	now := time.Now()
	timeOutstanding := now.Sub(p.SentAt)
	return timeOutstanding > time.Duration(config.Config.Timeout)*time.Second
}
