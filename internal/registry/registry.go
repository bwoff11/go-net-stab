package registry

import (
	"log"
	"time"

	"github.com/bwoff11/go-net-stab/internal/reporting"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Registry struct {
	Endpoints     map[int]string
	PingsSent     []Ping
	PingsRecieved []Ping
	HistorySize   int
	Sequence      int
	Connection    *icmp.PacketConn
}

func Create() Registry {
	conn, err := icmp.ListenPacket("ip4:icmp", "192.168.1.11")
	if err != nil {
		log.Fatal("Failed to listen for ICMP packets:", err)
	}

	return Registry{
		Endpoints:     make(map[int]string),
		PingsSent:     make([]Ping, 0),
		PingsRecieved: make([]Ping, 0),
		HistorySize:   100,
		Sequence:      0,
		Connection:    conn,
	}
}

func (r *Registry) AddEndpoint(endpoint string) {
	len := len(r.Endpoints)
	r.Endpoints[len] = endpoint
	log.Println("Added endpoint", endpoint, "with ID", len)
}

func (r *Registry) Run() {
	if err := r.StartListener(); err != nil {
		log.Fatal("Failed to start listener:", err)
	}

	go func() {
		for {
			log.Println("Sending pings for sequence", r.Sequence)
			for id, ip := range r.Endpoints {
				ping := CreatePing(id, r.Sequence, ip)
				ping.Send(r.Connection)
				r.PingsSent = append(r.PingsSent, ping)
			}
			r.Sequence++
			time.Sleep(3 * time.Second)
		}
	}()
}

func (r *Registry) StartListener() error {
	connection, err := icmp.ListenPacket("ip4:icmp", "192.168.1.11")
	if err != nil {
		return err
	}

	go func() {
		for {
			buffer := make([]byte, 1500)
			_, host, err := connection.ReadFrom(buffer)
			if err != nil {
				panic(err)
			}

			message, err := icmp.ParseMessage(1, buffer)
			if err != nil {
				panic(err)
			}

			switch message.Type {

			case ipv4.ICMPTypeEchoReply:
				r.HandleEchoReply(message, host.String())
			default:
				log.Println("Received unknown message type", message.Type)
			}
		}
	}()
	log.Println("Listener successfully started and waiting for ICMP messages")
	return nil
}

func (r *Registry) HandleEchoReply(message *icmp.Message, host string) {

	// Find the ping that matches the reply
	body := message.Body.(*icmp.Echo)
	id := int(body.ID)
	seq := int(body.Seq)
	ping := r.FindSentPing(id, seq)
	if ping == nil {
		log.Println("Received unexpected ping:", ping)
	}

	// Set the received time
	now := time.Now()
	ping.ReceivedAt = &now
	ping.RoundTripTime = float64(ping.CalculateRoundTripTime().Milliseconds())

	// Move from sent to received
	r.RemoveFromSent(*ping)
	r.PingsRecieved = append(r.PingsRecieved, *ping)

	// Update metrics
	reporting.RttGauge.With(
		prometheus.Labels{
			"source_ip":      ping.SourceIP,
			"destination_ip": ping.DestinationIP,
		},
	).Set(ping.RoundTripTime)
}

func (r *Registry) FindSentPing(id int, seq int) *Ping {
	for _, ping := range r.PingsSent {
		if ping.EndpointID == id && ping.Sequence == seq {
			return &ping
		}
	}
	return nil
}

func (r *Registry) RemoveFromSent(ping Ping) {
	for i, p := range r.PingsSent {
		if p.EndpointID == ping.EndpointID && p.Sequence == ping.Sequence {
			r.PingsSent = append(r.PingsSent[:i], r.PingsSent[i+1:]...)
			return
		}
	}
}
