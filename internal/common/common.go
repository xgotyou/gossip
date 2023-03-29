package common

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/tjarratt/babble"
	gossip "github.com/xgotyou/gossip/proto"
)

type Stream interface {
	Send(*gossip.Gossip) error
	Recv() (*gossip.Gossip, error)
}

func TellGossip(stream Stream, addr string, neighbourStates *sync.Map) {
	b := babble.NewBabbler()
	b.Separator = " "
	b.Count = 3

	for {
		msg := b.Babble()
		err := stream.Send(&gossip.Gossip{Text: msg})
		if err != nil {
			log.Printf("Can't send to gossip stream: %v", err)
			neighbourStates.Store(addr, false) // Mark client as inactive
			break
		}
		log.Printf("Sent: %s", msg)
		time.Sleep(time.Second)
	}
}

func HearGossip(stream Stream, addr string, neighbourStates *sync.Map) {
	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			log.Println("Gossip stream has been closed :'(")
			neighbourStates.Store(addr, false) // Mark client as inactive
			break
		}
		if err != nil {
			log.Printf("Error while receiving gossip stream from server: %v\n", err)
			neighbourStates.Store(addr, false) // Mark client as inactive
			break
		}

		log.Printf("Received: %s", recv.Text)
	}
}
