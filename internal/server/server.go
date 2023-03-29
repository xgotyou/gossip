package server

import (
	"log"
	"net"
	"sync"

	"github.com/xgotyou/gossip/internal/common"
	gossip "github.com/xgotyou/gossip/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Start(addr string) {
	lis, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("Can't listen for %v\n", addr)
	}

	log.Printf("Starting gRPC server on %v...\n", addr)

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	gossip.RegisterGossipServiceServer(grpcServer, GossipServer{})
	err = grpcServer.Serve(lis)

	if err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("Can't start gRPC server on %v\n", addr)
	}
}

type GossipServer struct {
	gossip.UnimplementedGossipServiceServer
}

func (s GossipServer) DiscussGossip(stream gossip.GossipService_DiscussGossipServer) error {
	go common.TellGossip(stream, "", &sync.Map{}) // TODO: Refactor this crunch
	common.HearGossip(stream, "", &sync.Map{})

	return nil
}
