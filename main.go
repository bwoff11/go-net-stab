package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Configuration struct {
	Interval  time.Duration
	Timeout   time.Duration
	Port      string
	Localhost string
	Endpoints []Endpoint
}

type Endpoint struct {
	Hostname string
	Address  string
	Location string
	Interval time.Duration
}

type PingService struct {
	Config     *Configuration
	Connection *icmp.PacketConn
	Sent       chan Ping
	Pending    []Ping
	Mutex      sync.Mutex
}

type Ping struct {
	ID     int
	Seq    int
	SentAt time.Time
}

var (
	SentPingsCounter *prometheus.CounterVec
	LostPingsCounter *prometheus.CounterVec
	RttGauge         *prometheus.GaugeVec
	pingService      *PingService
	configLocations  = []string{
		"/etc/go-net-stab/",
		"$HOME/.config/go-net-stab/",
		"$HOME/.go-net-stab",
		".",
	}
)

func init() {
	SentPingsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ping_sent_packet_total",
			Help: "Total number of packets sent",
		},
		[]string{
			"source_hostname",
			"destination_hostname",
			"destination_address",
			"destination_location",
		},
	)

	LostPingsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ping_lost_packet_total",
			Help: "Total number of packets lost",
		},
		[]string{
			"source_hostname",
			"destination_hostname",
			"destination_address",
			"destination_location",
		},
	)

	RttGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ping_rtt_milliseconds",
			Help: "Round trip time in milliseconds",
		},
		[]string{
			"source_hostname",
			"destination_hostname",
			"destination_address",
			"destination_location",
		},
	)

	prometheus.MustRegister(SentPingsCounter)
	prometheus.MustRegister(LostPingsCounter)
	prometheus.MustRegister(RttGauge)
}

func main() {
	pingService = &PingService{
		Config:  &Configuration{},
		Sent:    make(chan Ping),
		Pending: []Ping{},
	}

	if err := pingService.loadConfig(); err != nil {
		log.Fatal(err)
	}

	if err := pingService.createConnection(); err != nil {
		log.Fatal(err)
	}

	pingService.startPingingEndpoints()

	go pingService.appendPendingPings()
	go pingService.createListener()
	go pingService.checkForLostPings()

	http.Handle("/metrics", promhttp.Handler())

	metricsURL := fmt.Sprintf("http://localhost:%s/metrics", pingService.Config.Port)
	log.Printf("Metrics are being exposed at: %s", metricsURL)

	log.Fatal(http.ListenAndServe(":"+pingService.Config.Port, nil))
}

func (ps *PingService) loadConfig() error {
	viper.SetConfigName("config")

	for _, location := range configLocations {
		viper.AddConfigPath(location)
	}

	if err := viper.ReadInConfig(); err != nil {
		return errors.New("Fatal error config file: " + err.Error())
	}

	if err := viper.Unmarshal(ps.Config); err != nil {
		return errors.New("Fatal error config file: " + err.Error())
	}

	ps.Config.Interval *= time.Millisecond
	ps.Config.Timeout *= time.Millisecond

	return nil
}

func (ps *PingService) createConnection() error {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}

	ps.Connection = conn
	return nil
}

func (ps *PingService) startPingingEndpoints() {
	var wg sync.WaitGroup
	for i, endpoint := range ps.Config.Endpoints {
		wg.Add(1)
		go func(id int, endpoint Endpoint) {
			defer wg.Done()
			ps.pingEndpoint(id, endpoint)
		}(i, endpoint)
	}
}

func (ps *PingService) pingEndpoint(id int, endpoint Endpoint) {
	sequence := 0
	ticker := time.NewTicker(ps.Config.Interval)

	for range ticker.C {
		m := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   id,
				Seq:  sequence,
				Data: []byte("we've been trying to reach you about your car's extended warranty"),
			},
		}

		b, err := m.Marshal(nil)
		if err != nil {
			log.Fatal(err)
		}

		_, err = ps.Connection.WriteTo(b, &net.IPAddr{IP: net.ParseIP(endpoint.Address)})
		if err != nil {
			log.Printf("Error sending ping to %s at %s in %s: %v", endpoint.Hostname, endpoint.Address, endpoint.Location, err)
			break
		}

		ps.Sent <- Ping{
			ID:     id,
			Seq:    sequence,
			SentAt: time.Now(),
		}

		SentPingsCounter.WithLabelValues(
			ps.Config.Localhost,
			endpoint.Hostname,
			endpoint.Address,
			endpoint.Location,
		).Inc()

		sequence++
	}
}

func (ps *PingService) appendPendingPings() {
	for {
		ps.Pending = append(ps.Pending, <-ps.Sent)
	}
}

func (ps *PingService) createListener() {
	rb := make([]byte, 1500)

	for {
		n, _, err := ps.Connection.ReadFrom(rb)
		if err != nil {
			log.Printf("Error reading ICMP reply: %v", err)
			continue
		}

		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			log.Printf("Error parsing ICMP message: %v", err)
			continue
		}

		if rm.Type == ipv4.ICMPTypeEchoReply {
			ps.processPingReply(rm)
		}
	}
}

func (ps *PingService) processPingReply(rm *icmp.Message) {
	ps.Mutex.Lock()
	defer ps.Mutex.Unlock()

	for i, ping := range ps.Pending {
		if ping.ID == rm.Body.(*icmp.Echo).ID && ping.Seq == rm.Body.(*icmp.Echo).Seq {
			ps.Pending = append(ps.Pending[:i], ps.Pending[i+1:]...)

			rtt := float64(time.Since(ping.SentAt).Milliseconds())

			RttGauge.WithLabelValues(
				ps.Config.Localhost,
				ps.Config.Endpoints[ping.ID].Hostname,
				ps.Config.Endpoints[ping.ID].Address,
				ps.Config.Endpoints[ping.ID].Location,
			).Set(rtt)

			break
		}
	}
}

func (ps *PingService) checkForLostPings() {
	ticker := time.NewTicker(ps.Config.Timeout)

	for range ticker.C {
		ps.Mutex.Lock()

		for i, ping := range ps.Pending {
			if time.Since(ping.SentAt) > ps.Config.Timeout {
				ps.Pending = append(ps.Pending[:i], ps.Pending[i+1:]...)

				LostPingsCounter.WithLabelValues(
					ps.Config.Localhost,
					ps.Config.Endpoints[ping.ID].Hostname,
					ps.Config.Endpoints[ping.ID].Address,
					ps.Config.Endpoints[ping.ID].Location,
				).Inc()
			}
		}

		ps.Mutex.Unlock()
	}
}
