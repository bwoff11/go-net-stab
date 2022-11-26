package registry

import (
	"log"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Registry struct {
	Endpoints     []string
	PingsSent     []Ping
	PingsRecieved []Ping
	HistorySize   int
	Connection    *icmp.PacketConn
}

func Create() Registry {
	conn, err := icmp.ListenPacket("ip4:icmp", "192.168.1.11")
	if err != nil {
		log.Fatal("Failed to listen for ICMP packets:", err)
	}

	return Registry{
		PingsSent:     make([]Ping, 0),
		PingsRecieved: make([]Ping, 0),
		HistorySize:   100,
		Connection:    conn,
	}
}

func (r *Registry) AddEndpoint(endpoint string) {
	r.Endpoints = append(r.Endpoints, endpoint)
}

func (r *Registry) Run() {
	if err := r.StartListener(); err != nil {
		log.Fatal("Failed to start listener:", err)
	}

	go func() {
		for {
			for _, endpoint := range r.Endpoints {
				ping := CreatePing(endpoint)
				ping.Send(r.Connection, 1)
				ping.LogData()
				r.PingsSent = append(r.PingsSent, ping)
			}
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
			_, _, err := connection.ReadFrom(buffer)
			if err != nil {
				panic(err)
			}

			message, err := icmp.ParseMessage(1, buffer)
			if err != nil {
				panic(err)
			}

			switch message.Type {

			case ipv4.ICMPTypeEchoReply:
				//body := message.Body.(*icmp.Echo)
			default:
				log.Println("Received unknown message type", message.Type)
			}
		}
	}()
	log.Println("Listener successfully started and waiting for ICMP messages")
	return nil
}
