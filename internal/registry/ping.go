package registry

import (
	"log"
	"net"
	"time"

	"github.com/bwoff11/go-net-stab/internal/config"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Ping struct {
	EndpointID    int
	Sequence      int
	SourceIP      string
	DestinationIP string
	SentAt        time.Time
	ReceivedAt    *time.Time
	RoundTripTime float64
}

func (p *Ping) Send() error {

	// Create ICMP message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   p.EndpointID,
			Seq:  p.Sequence,
			Data: []byte("We've been trying to reach you about your car's extended warranty"),
		},
	}

	// Marshal ICMP message
	bytes, err := msg.Marshal(nil)
	if err != nil {
		log.Fatalln("Error marshalling payload:", err)
		return nil
	}

	// Send ICMP message
	if _, err := registry.Connection.WriteTo(bytes, &net.IPAddr{IP: net.ParseIP(p.DestinationIP)}); err != nil {
		return err
	}
	p.SentAt = time.Now()

	return nil
}

func (p *Ping) SetAsRecieved() {
	now := time.Now()
	p.ReceivedAt = &now

	rtt := now.Sub(p.SentAt)
	p.RoundTripTime = float64(rtt.Milliseconds())
}

// Checks to see if the time since the ping was sent exceeds the configured timeout
func (p *Ping) IsLost() bool {
	if p.ReceivedAt != nil {
		return false
	}
	now := time.Now()
	timeOutstanding := now.Sub(p.SentAt)
	return timeOutstanding > time.Duration(config.Config.Timeout)*time.Second
}
