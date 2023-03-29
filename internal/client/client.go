package client

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/xgotyou/gossip/internal/common"
	gossip "github.com/xgotyou/gossip/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func startGossiping(ctx context.Context, c gossip.GossipServiceClient, addr string, neighbourStates *sync.Map) {
	stream, err := c.DiscussGossip(ctx)
	if err != nil {
		log.Printf("Can't start gossip stream: %v", err)
		return
	}

	neighbourStates.Store(addr, true) // mark client as active

	go common.TellGossip(stream, addr, neighbourStates)
	go common.HearGossip(stream, addr, neighbourStates)
}

func Start(addr string, neighbourStates *sync.Map) {
	cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Printf("Can't dial %s using gRPC: %v", addr, err)
		return
	}

	defer cc.Close()

	c := gossip.NewGossipServiceClient(cc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go startGossiping(ctx, c, addr, neighbourStates)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)
	<-done
}

func Manage(neighbours []string, pollInt time.Duration) {
	neighbourStates := sync.Map{}

	for _, nb := range neighbours {
		neighbourStates.Store(nb, false)
	}

	for {
		neighbourStates.Range(func(addr any, running any) bool {
			if !running.(bool) {
				go Start(addr.(string), &neighbourStates)
			}
			return true
		})
		time.Sleep(pollInt)
	}
}
