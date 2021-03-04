// Package main implements a simple gRPC server that implements the esbbridge rpc server described in esbbridge_rpc.proto
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/spritkopf/esb-bridge/pkg/server/service"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type esbBridgeServer struct {
	pb.UnimplementedEsbBridgeServer
}

// GetFeature returns the feature at the given point.
func (s *esbBridgeServer) Transfer(ctx context.Context, msg *pb.EsbMessage) (*pb.EsbMessage, error) {

	// simple echo for now
	return &pb.EsbMessage{Addr: msg.Addr, Cmd: msg.Cmd, Payload: msg.Payload}, nil
}

// ListFeatures lists all features contained within the given bounding Rectangle.
func (s *esbBridgeServer) Listen(listener *pb.Listener, messageStream pb.EsbBridge_ListenServer) error {

	// if err := messageStream.Send(&pb.EsbMessage{}); err != nil {
	// 	return err
	// }

	return nil
}

func newServer() *esbBridgeServer {
	s := &esbBridgeServer{}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	log.Printf("Serving on port %v\n", *port)
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEsbBridgeServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
