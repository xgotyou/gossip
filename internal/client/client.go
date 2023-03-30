package client

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/xgotyou/gossip/internal/common"
	gossip "github.com/xgotyou/gossip/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Start(ctx context.Context, addr string, neighbourStates *sync.Map) {
	cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Can't dial %s using gRPC: %v", addr, err)
		return
	}
	defer cc.Close()

	c := gossip.NewGossipServiceClient(cc)
	stream, err := c.DiscussGossip(ctx)
	if err != nil {
		log.Printf("Can't start gossip stream: %v", err)
		return
	}

	neighbourStates.Store(addr, true) // mark client as active

	go common.TellGossip(stream, addr, neighbourStates)
	go common.HearGossip(stream, addr, neighbourStates)

	<-ctx.Done()
}

func Manage(ctx context.Context, neighbours []string, pollInt time.Duration) {
	neighbourStates := sync.Map{}

	for _, nb := range neighbours {
		neighbourStates.Store(nb, false)
	}

	for {
		neighbourStates.Range(func(addr any, running any) bool {
			if !running.(bool) {
				go Start(ctx, addr.(string), &neighbourStates)
			}
			return true
		})
		time.Sleep(pollInt)
	}
}
