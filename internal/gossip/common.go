package gossip

import (
	"log"
	"time"

	"github.com/tjarratt/babble"
	gossippb "github.com/xgotyou/gossip/proto"
)

type Stream interface {
	Send(*gossippb.Gossip) error
	Recv() (*gossippb.Gossip, error)
}

func TellGossip(stream Stream, addr string, m *NeighbourManager) {
	b := babble.NewBabbler()
	b.Separator = " "
	b.Count = 3

	for {
		msg := b.Babble()
		err := stream.Send(&gossippb.Gossip{Text: msg})
		if err != nil {
			log.Printf("Can't send to gossip stream to %v: %v", addr, err)
			if m != nil {
				m.Remove(addr)
			}
			break
		}
		log.Printf("Sent to %v: %v", addr, msg)
		time.Sleep(time.Second)
	}
}

func HearGossip(stream Stream, addr string, m *NeighbourManager) {
	for {
		recv, err := stream.Recv()
		if err != nil {
			log.Printf("Neighbour %v became inactive :'(", addr)
			if m != nil {
				m.Remove(addr)
			}
			break
		}

		log.Printf("Received from %v: %v", addr, recv.Text)
	}
}
