package main

import (
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Pinger struct {
	ID            int
	SourceIP      net.IP
	DestinationIP net.IP
	Sequence      int
	Connection    *icmp.PacketConn
}

func (p *Pinger) Run(interval int) {
	p.CreateConnection()
	for {
		payload := p.CreatePayload()
		p.SendPing(payload)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func (p *Pinger) CreateConnection() {
	conn, err := icmp.ListenPacket("ip4:icmp", "192.168.1.11")
	if err != nil {
		panic(err)
	}
	p.Connection = conn
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
		panic(err)
	} else {
		p.Sequence++
		return bytes
	}
}

func (p *Pinger) SendPing(payload []byte) {
	newPing := Ping{
		PingerID:      p.ID,
		SourceIP:      p.SourceIP,
		DestinationIP: p.DestinationIP,
		Sequence:      p.Sequence,
		SentAt:        time.Now(),
		ReceivedAt:    nil,
	}
	sentPings <- newPing
	if _, err := p.Connection.WriteTo(
		payload,
		&net.IPAddr{
			IP: newPing.DestinationIP,
		}); err != nil {
		panic(err)
	}
}
