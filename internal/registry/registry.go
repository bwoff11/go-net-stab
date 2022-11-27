package registry

import (
	"log"
	"time"

	"github.com/bwoff11/go-net-stab/internal/config"
	"github.com/bwoff11/go-net-stab/internal/reporting"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Registry struct {
	Endpoints    map[int]string
	PendingPings []Ping
	Sequence     int
	Connection   *icmp.PacketConn
}

var registry *Registry

func Create() {

	// Create connection for outgoing pings
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatal("Failed to listen for ICMP packets:", err)
	}

	// Initialize registry
	registry = &Registry{
		Endpoints:    make(map[int]string),
		PendingPings: make([]Ping, 0),
		Sequence:     0,
		Connection:   conn,
	}

	// Add endpoints
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
			for _, ping := range r.PendingPings {
				if ping.IsLost() {
					log.Println("Ping", ping.Sequence, "to", ping.DestinationIP, "is lost")

					// Remove ping from pending list
					r.RemovePingFromPending(&ping)

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

				// Create ping
				ping := Ping{
					EndpointID:    id,
					Sequence:      r.Sequence,
					SourceIP:      registry.Connection.LocalAddr().String(),
					DestinationIP: registry.Endpoints[id],
				}

				// Send ping
				ping.Send()

				// Add ping to pending list
				r.PendingPings = append(r.PendingPings, ping)

				// Update metrics
				reporting.SentPingsCounter.With(
					prometheus.Labels{
						"source_ip":      ping.SourceIP,
						"destination_ip": ping.DestinationIP,
					},
				).Inc()
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

	// Remove ping from pending list
	r.RemovePingFromPending(ping)

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

	for _, ping := range r.PendingPings {
		if ping.EndpointID == id && ping.Sequence == seq {
			return &ping
		}
	}
	return nil
}

func (r *Registry) RemovePingFromPending(ping *Ping) {
	for i, p := range r.PendingPings {
		if p.Sequence == ping.Sequence && p.EndpointID == ping.EndpointID {
			r.PendingPings = append(r.PendingPings[:i], r.PendingPings[i+1:]...)
			return
		}
	}
}
