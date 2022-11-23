package listener

import (
	"log"

	"github.com/bwoff11/go-net-stab/internal/registry"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func Start() error {
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
				body := message.Body.(*icmp.Echo)
				registry.HandleEchoReply(body.ID, body.Seq, host)

			case ipv4.ICMPTypeDestinationUnreachable:
				//handleDestinationUnreachableReply(message, host.(*net.IPAddr))
			default:
				log.Println("Received unknown message type", message.Type)
			}
		}
	}()
	log.Println("Listener successfully started and waiting for ICMP messages")
	return nil
}

/*
func handleEchoReply(message *icmp.Message, host *net.IPAddr) {
	// TODO: Add check for reply to timed-out packet

	body := message.Body.(*icmp.Echo)
	log.Printf("Received response from %s with id %d and sequence %d", host, body.ID, body.Seq)

	// Search outstanding pings for a match
	for _, ping := range outstandingPings {
		if ping.PingerID == body.ID && ping.Sequence == body.Seq {
			//handlePingMatch(ping)
			return
		}
	}
}*/

/*
func handlePingMatch(ping Ping) {
	now := time.Now()
	ping.ReceivedAt = &now
	rtt := float64(ping.CalculateRoundTripTime().Milliseconds())

	reporting.RttGauge.With(
		prometheus.Labels{
			"source_ip":      ping.SourceIP,
			"destination_ip": ping.DestinationIP,
		},
	).Set(rtt)
	log.Println("RTT:", rtt)

	// Remove ping from outstanding pings
	for i, p := range outstandingPings {
		if p == ping {
			outstandingPings = append(outstandingPings[:i], outstandingPings[i+1:]...)
			return
		}
	}
}*/
