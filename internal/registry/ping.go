package registry

import "time"

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
