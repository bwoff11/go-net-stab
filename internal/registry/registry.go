package registry

import (
	"errors"
	"log"
	"time"

	"github.com/bwoff11/go-net-stab/internal/config"
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

var registry *Registry

func Create() {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatal("Failed to listen for ICMP packets:", err)
	}

	registry = &Registry{
		Endpoints:     make(map[int]string),
		PingsSent:     make([]Ping, 0),
		PingsRecieved: make([]Ping, 0),
		HistorySize:   100,
		Sequence:      0,
		Connection:    conn,
	}

	for _, endpoint := range config.Config.Endpoints {
		registry.AddEndpoint(endpoint)
	}
}

func Start() {
	if err := registry.StartListener(); err != nil {
		log.Fatal("Failed to start listener:", err)
	}
	if err := registry.StartEndpointPinger(); err != nil {
		log.Fatal("Failed to start endpoint pinger:", err)
	}
	if err := registry.StartLostPingWatcher(); err != nil {
		log.Fatal("Failed to start lost ping watcher:", err)
	}
}

func (r *Registry) AddEndpoint(endpoint string) {
	len := len(r.Endpoints)
	r.Endpoints[len] = endpoint
	log.Println("Added endpoint", endpoint, "with ID", len)
}

func (r *Registry) StartLostPingWatcher() error {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			for _, ping := range r.PingsSent {
				if ping.IsLost() {
					log.Println("Ping", ping.Sequence, "to", ping.DestinationIP, "is lost")
					r.RemovePingFromSent(ping)

					// Update metrics
					reporting.LostPingsCounter.With(
						prometheus.Labels{
							"source_ip":      ping.SourceIP,
							"destination_ip": ping.DestinationIP,
						},
					).Inc()
				}
			}
		}
	}()
	return nil
}

func (r *Registry) StartEndpointPinger() error {
	interval := config.Config.Interval
	go func() {
		for {
			log.Println("Sending pings for sequence", r.Sequence)
			for id := range r.Endpoints {
				ping := CreatePing(id, r.Sequence)
				ping.Send()
				r.PingsSent = append(r.PingsSent, ping)
				log.Println("Sent ping", ping.Sequence, "to", ping.DestinationIP)
			}
			r.Sequence++
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()
	return nil
}

func (r *Registry) StartListener() error {
	connection, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
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
				//log.Println("Received unknown message type", message.Type)
			}
		}
	}()
	log.Println("Listener successfully started and waiting for ICMP messages")
	return nil
}

func (r *Registry) HandleEchoReply(message *icmp.Message, host string) {
	ping := r.MatchReplyMessageToSentPing(message)
	if ping == nil {
		log.Println("Received unexpected ping:", ping)
	}
	ping.SetAsRecieved()
	r.MoveFromSentToRecieved(*ping)

	// Update metrics
	reporting.RttGauge.With(
		prometheus.Labels{
			"source_ip":      ping.SourceIP,
			"destination_ip": ping.DestinationIP,
		},
	).Set(ping.RoundTripTime)
}

func (r *Registry) MatchReplyMessageToSentPing(message *icmp.Message) *Ping {
	body := message.Body.(*icmp.Echo)
	id := int(body.ID)
	seq := int(body.Seq)

	for _, ping := range r.PingsSent {
		if ping.EndpointID == id && ping.Sequence == seq {
			return &ping
		}
	}
	return nil
}

func (r *Registry) MoveFromSentToRecieved(ping Ping) error {
	if err := r.RemovePingFromSent(ping); err != nil {
		return err
	}
	if err := r.AddPingToRecieved(ping); err != nil {
		return err
	}
	return nil
}

func (r *Registry) RemovePingFromSent(p Ping) error {
	for i, ping := range r.PingsSent {
		if ping.EndpointID == p.EndpointID && ping.Sequence == p.Sequence {
			r.PingsSent = append(r.PingsSent[:i], r.PingsSent[i+1:]...)
			return nil
		}
	}
	return errors.New("could not find ping to remove")
}

func (r *Registry) AddPingToRecieved(p Ping) error {
	if len(r.PingsRecieved) == r.HistorySize {
		r.PingsRecieved = r.PingsRecieved[1:]
	}
	r.PingsRecieved = append(r.PingsRecieved, p)
	return nil
}
