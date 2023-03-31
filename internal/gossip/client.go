package gossip

import (
	"context"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	gossippb "github.com/xgotyou/gossip/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NeighbourManager struct {
	localIP    string
	pollInt    time.Duration
	mtx        sync.Mutex
	neighbours map[string]bool
}

func (m *NeighbourManager) Call(ctx context.Context, localIP string, pollInt time.Duration) {
	m.localIP = localIP
	m.pollInt = pollInt
	m.neighbours = make(map[string]bool)

	go m.discoverNeighbours()
	go m.manage(ctx)
}

func (m *NeighbourManager) markActive(neighbourAddr string) {
	m.mtx.Lock()
	m.neighbours[neighbourAddr] = true
	m.mtx.Unlock()
}

func (m *NeighbourManager) add(neighbourAddr string) {
	m.mtx.Lock()
	_, exists := m.neighbours[neighbourAddr]
	if !exists {
		m.neighbours[neighbourAddr] = false
		log.Println("A new neighbour to gossip with:", neighbourAddr)
	}
	m.mtx.Unlock()
}

func (m *NeighbourManager) Remove(neighbourAddr string) {
	m.mtx.Lock()
	delete(m.neighbours, neighbourAddr)
	m.mtx.Unlock()
}

func (m *NeighbourManager) discoverNeighbours() {
	pc, err := net.ListenPacket("udp4", ":8831")
	if err != nil {
		log.Println("Can't resolve UDP address to monitor neighbours", err)
	}

	msg := make([]byte, 1024)

	for {
		_, addr, err := pc.ReadFrom(msg)
		if err != nil {
			log.Println("Error while reading UDP message:", err)
		}

		message := string(msg)
		if !strings.Contains(message, m.localIP) {
			log.Println("Neigbour contact received:", string(msg), addr)
			_, remoteAddr, found := strings.Cut(message, "gossip://")
			if found {
				m.add(strings.TrimRight(remoteAddr, "\r\n\x00"))
			}
		}
	}
}

func (m *NeighbourManager) startConversation(ctx context.Context, addr string) {
	cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Can't dial %v using gRPC: %v", addr, err)
		return
	}
	defer cc.Close()

	c := gossippb.NewGossipServiceClient(cc)
	stream, err := c.DiscussGossip(ctx)
	if err != nil {
		log.Printf("Can't start gossip stream with %v: %v", addr, err)
		return
	}

	m.markActive(addr)

	go TellGossip(stream, addr, m)
	go HearGossip(stream, addr, m)

	<-ctx.Done()
}

func (m *NeighbourManager) manage(ctx context.Context) {
	for {
		for addr, talking := range m.neighbours {
			if !talking {
				go m.startConversation(ctx, addr)
			}
		}
		time.Sleep(m.pollInt)
	}
}
