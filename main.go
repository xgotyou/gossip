package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/xgotyou/gossip/internal/gossip"
)

type config struct {
	Port    int           `env:"PORT" env-default:"5051"`
	PollInt time.Duration `env:"POLLINT" env-default:"5s"`
}

func main() {
	// Load configuration
	var cfg config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatalf("Error reading configuration: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	localIP, broadcastIP, err := fetchIPs()
	if err != nil {
		log.Fatalf("Can't fetch IPs: %v", err)
	}
	log.Printf("Start broadcasting presence on %v...", localIP)

	go broadcastPresence(ctx, localIP, cfg.Port, broadcastIP, cfg.PollInt)

	// Start gPRC server
	go gossip.StartServer("0.0.0.0:" + strconv.Itoa(cfg.Port))

	// Manage connections with other services
	m := &gossip.NeighbourManager{}
	go m.Call(ctx, localIP.String(), cfg.PollInt)

	// Wait for external interuption signal and shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	log.Println("Service terminated.")
}

// Once in pollInt it broadcasts the message with gRPC server address (localIp + port)
// in a form of "gossip://1.2.3.4:1234" so that the receiver can distinguish it from
// other broadcasted messages and extract all the info needed to establish a connection.
func broadcastPresence(ctx context.Context, localIP net.IP, port int, broadcastIP net.IP, pollInt time.Duration) {
	remote, err := net.ResolveUDPAddr("udp4", broadcastIP.String()+":8831")
	if err != nil {
		log.Fatalf("Can't resolve UDP broadcast address: %v", err)
	}
	conn, err := net.DialUDP("udp4", nil, remote)
	if err != nil {
		log.Fatalf("Can't dial UDP connection: %v", err)
	}
	defer conn.Close()

	t := time.NewTicker(pollInt)
OUTER:
	for {
		select {
		case <-t.C:
			_, err = conn.Write([]byte("gossip://" + localIP.String() + ":" + strconv.Itoa(port)))
			if err != nil {
				log.Printf("Can't broadcast service's presence to the network: %v", err)
			}
		case <-ctx.Done():
			log.Println("Presence broadcasting stopped")
			break OUTER
		}
	}
}

func fetchIPs() (localIP net.IP, broadcastIP net.IP, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalln(err)
	}

	var broadcastIPs []net.IP
	var localIPs []net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			netID := ipnet.IP.Mask(ipnet.Mask)
			broadcast := make(net.IP, len(ipnet.Mask))
			for i := range ipnet.Mask {
				broadcast[i] = netID[i] | ^ipnet.Mask[i]
			}
			broadcastIPs = append(broadcastIPs, broadcast)
			localIPs = append(localIPs, ipnet.IP.To4())
		}
	}
	if len(broadcastIPs) < 1 || len(localIPs) < 1 {
		return nil, nil, errors.New("available IPs not found, probably not connected to IPv4 networks")
	}

	return localIPs[0], broadcastIPs[0], nil
}
