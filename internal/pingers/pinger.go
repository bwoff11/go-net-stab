package pingers

import (
	"log"
	"net"
	"time"

	"github.com/bwoff11/go-net-stab/internal/config"
	"github.com/bwoff11/go-net-stab/internal/registry"
	"github.com/bwoff11/go-net-stab/internal/reporting"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Pinger struct {
	ID            int
	SourceIP      string
	DestinationIP string
	Sequence      int
	Connection    *icmp.PacketConn
}

func (p *Pinger) Start() {
	conn, err := icmp.ListenPacket("ip4:icmp", "192.168.1.11")
	if err != nil {
		log.Fatal("Error creating connection:", err)
	}
	p.Connection = conn

	log.Println("Starting pinger", p.ID, "for", p.DestinationIP)

	for {
		payload := p.CreatePayload()
		p.SendPing(payload)
		p.Sequence++
		time.Sleep(time.Duration(config.Config.Interval) * time.Second)
	}
}

func (p *Pinger) CreatePayload() []byte {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   p.ID,
			Seq:  p.Sequence,
			Data: []byte("We've been trying to reach you about your car's extended warranty"),
		},
	}
	if bytes, err := msg.Marshal(nil); err != nil {
		log.Fatalln("Error marshalling payload:", err)
		return nil
	} else {
		return bytes
	}
}

func (p *Pinger) SendPing(payload []byte) {
	newPing := registry.Ping{
		PingerID:      p.ID,
		SourceIP:      p.SourceIP,
		DestinationIP: p.DestinationIP,
		Sequence:      p.Sequence,
		SentAt:        time.Now(),
		ReceivedAt:    nil,
	}
	if _, err := p.Connection.WriteTo(
		payload,
		&net.IPAddr{
			IP: net.ParseIP(p.DestinationIP),
		}); err != nil {
		log.Println("Error sending ping:", err)
	}
	registry.SentPings <- newPing
	reporting.SentPacketCounter.With(
		prometheus.Labels{
			"source_ip":      p.SourceIP,
			"destination_ip": p.DestinationIP,
		},
	).Inc()
}
