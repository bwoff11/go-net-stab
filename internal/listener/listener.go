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
				registry.HandleDestinationUnreachableReply(host)
			case ipv4.ICMPTypeTimeExceeded:
				registry.HandleTimeExceededReply(host)
			default:
				log.Println("Received unknown message type", message.Type)
				body := message.Body.(*icmp.Echo)
				log.Println("ID:", body.ID)
				log.Println("Seq:", body.Seq)
				log.Println("Data:", body.Data)
			}
		}
	}()
	log.Println("Listener successfully started and waiting for ICMP messages")
	return nil
}
