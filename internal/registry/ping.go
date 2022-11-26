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
	Sequence      int
	SourceIP      string
	DestinationIP string
	SentAt        time.Time
	ReceivedAt    *time.Time
}

func CreatePing(destinationIP string) Ping {
	return Ping{
		SourceIP:      "192.168.1.11",
		DestinationIP: destinationIP,
	}
}

func (p *Ping) LogData() {
	log.Println("Sent ping:", p)
}

func (p *Ping) Send(conn *icmp.PacketConn, sequence int) error {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   0,
			Seq:  p.Sequence,
			Data: []byte("We've been trying to reach you about your car's extended warranty"),
		},
	}
	bytes, err := msg.Marshal(nil)
	if err != nil {
		log.Fatalln("Error marshalling payload:", err)
		return nil
	}

	p.SentAt = time.Now()
	if _, err := conn.WriteTo(bytes, &net.IPAddr{IP: net.ParseIP(p.DestinationIP)}); err != nil {
		return err
	}
	return nil
}

func (p *Ping) SetReceived() {
	now := time.Now()
	p.ReceivedAt = &now
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
