// Package main implements a simple gRPC server that implements the esbbridge rpc server described in esbbridge_rpc.proto
package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/spritkopf/esb-bridge/pkg/esbbridge"
	pb "github.com/spritkopf/esb-bridge/pkg/server/service"
)

var (
	port   = flag.Int("port", 10000, "The server port")
	device = flag.String("device", "/dev/ttyACM0", "The esbbridge serial device")
)

type esbBridgeServer struct {
	pb.UnimplementedEsbBridgeServer
}

// GetFeature returns the feature at the given point.
func (s *esbBridgeServer) Transfer(ctx context.Context, msg *pb.EsbMessage) (*pb.EsbMessage, error) {

	log.Printf("Transfer Message: %v\n", msg)

	answer, err := esbbridge.Transfer(esbbridge.EsbMessage{Address: msg.Addr, Cmd: msg.Cmd[0], Payload: msg.Payload})

	if err != nil {
		log.Printf("Transfer error: %v", err)
	}

	return &pb.EsbMessage{Addr: msg.Addr, Cmd: []byte{answer.Cmd}, Error: []byte{answer.Error}, Payload: answer.Payload}, nil
}

// Listen starts to listen for a specific messages and streams incoming messages to the client
func (s *esbBridgeServer) Listen(listener *pb.Listener, messageStream pb.EsbBridge_ListenServer) error {

	log.Printf("Attach listener for Address: %v, Command %v", listener.Addr, listener.Cmd)
	streamDone := messageStream.Context().Done()

	listenAddr := [5]byte{}
	copy(listenAddr[:5], listener.Addr)

	lc := make(chan esbbridge.EsbMessage, 1)
	esbbridge.AddListener(listenAddr, listener.Cmd[0], lc)

listenLoop:
	for {
		select {
		case msg := <-lc:
			log.Printf("Incoming Message: %v\n", msg)
			err := messageStream.Send(&pb.EsbMessage{Addr: msg.Address, Cmd: []byte{msg.Cmd}, Payload: msg.Payload})
			if err != nil {
				return err
			}
		case <-streamDone:
			log.Printf("Listener %v, %v canceled by client", listener.Addr, listener.Cmd)
			log.Printf("Detach listener for Address: %v, Command %v", listener.Addr, listener.Cmd)
			esbbridge.RemoveListener(lc)
			break listenLoop
		}
	}

	log.Println("Done listening")

	return nil
}

func newServer() *esbBridgeServer {
	s := &esbBridgeServer{}
	return s
}

// Start starts the esb-bridge RPC server in a goroutine. To cancel the execution,
// call the returned cancel function
// Params:
//   device: device string to connect to (e.g. /dev/ttyACM0)
//   port: TCP port for the RPC server
func Start(device string, port uint) (context.CancelFunc, error) {

	err := esbbridge.Open(device)
	if err != nil {
		log.Printf("Could not open connection to esb-bridge device: %v", err)
		return nil, err
	}
	fwVersion, err := esbbridge.GetFwVersion()
	if err != nil {
		log.Printf("Error reading Firmware version of esb-bridge device: %v", err)
		return nil, err
	}
	log.Printf("esb-bridge firmware version: %v", fwVersion)

	ctx, cancel := context.WithCancel(context.Background())
	go func(context.Context) {
		defer esbbridge.Close()

		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		var opts []grpc.ServerOption

		log.Printf("Serving on port %v\n", port)
		grpcServer := grpc.NewServer(opts...)
		pb.RegisterEsbBridgeServer(grpcServer, newServer())
		grpcServer.Serve(lis)
	}(ctx)

	return cancel, nil
}
