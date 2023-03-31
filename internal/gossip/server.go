package gossip

import (
	"log"
	"net"

	gossippb "github.com/xgotyou/gossip/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
)

func StartServer(addr string) {
	lis, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("Can't listen for %v\n", addr)
	}

	log.Printf("Starting gRPC server on %v...\n", addr)

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	gossippb.RegisterGossipServiceServer(grpcServer, GossipServer{})
	err = grpcServer.Serve(lis)

	if err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("Can't start gRPC server on %v\n", addr)
	}
}

type GossipServer struct {
	gossippb.UnimplementedGossipServiceServer
}

func (s GossipServer) DiscussGossip(stream gossippb.GossipService_DiscussGossipServer) error {
	peer, _ := peer.FromContext(stream.Context())
	addr := peer.Addr.String()
	go TellGossip(stream, addr, nil) // TODO: A bit cruncy, would be nice to refactor
	HearGossip(stream, addr, nil)

	return nil
}
