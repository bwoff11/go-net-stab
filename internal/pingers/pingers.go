package pingers

import (
	"log"
	"net"
	"time"

	"github.com/bwoff11/go-net-stab/internal/config"
	"github.com/bwoff11/go-net-stab/internal/registry"
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

var pingers []Pinger

func Start() error {
	createPingers()
	for _, pinger := range pingers {
		go pinger.Start()
	}
	log.Println("All pingers successfully created and started")
	return nil
}

func createPingers() {
	var id int
	for _, endpoint := range config.Config.Endpoints {
		pingers = append(pingers, Pinger{
			ID:            id,
			SourceIP:      "192.168.1.11",
			DestinationIP: endpoint,
		})
		log.Println("Created new pinger for", endpoint, "with ID", id)
		id++
	}
}

func (p *Pinger) Start() {
	conn, err := icmp.ListenPacket("ip4:icmp", "192.168.1.11")
	if err != nil {
		panic(err)
	}
	p.Connection = conn

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
		panic(err)
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
		panic(err)
	}
	registry.SentPings <- newPing
}
