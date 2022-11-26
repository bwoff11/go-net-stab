package registry

import (
	"log"
	"time"

	"golang.org/x/net/icmp"
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
	//defer conn.Close()

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
	go func() {
		for {
			for _, endpoint := range r.Endpoints {
				ping := CreatePing(endpoint)
				ping.LogData()
				ping.Send(r.Connection, 1)
				r.PingsSent = append(r.PingsSent, ping)
			}
			time.Sleep(3 * time.Second)
		}
	}()
}
