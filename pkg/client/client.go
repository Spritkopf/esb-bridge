// Package main implements a simple gRPC client that implements the esbbridge rpc client described in esbbridge_rpc.proto
package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	pb "github.com/spritkopf/esb-bridge/pkg/server/service"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
)

// transfer sends a message to a peripheral device and returns the answer message
func transfer(client pb.EsbBridgeClient, msg *pb.EsbMessage) {
	log.Printf("Sending Message %v", msg)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	answerMessage, err := client.Transfer(ctx, msg)
	if err != nil {
		log.Fatalf("%v.Transfer(_) = _, %v: ", client, err)
	}
	log.Printf("Answer: %v\n", answerMessage)
}

// printFeatures lists all the features within the given bounding Rectangle.
func listen(client pb.EsbBridgeClient, listener *pb.Listener) {
	log.Printf("Start listening: %v", listener)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.Listen(ctx, listener)
	if err != nil {
		log.Fatalf("%v.Listen(_) = _, %v", client, err)
	}
	for {
		incomingMessage, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}
		log.Printf("Incoming Message: %v", incomingMessage)
	}
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewEsbBridgeClient(conn)

	// Test transfer function
	transfer(client, &pb.EsbMessage{Addr: []byte{1, 2, 3, 4, 5}, Cmd: []byte{128}, Payload: []byte{9, 8, 7}})

	listen(client, &pb.Listener{Addr: []byte{12, 13, 14, 15, 16}, Cmd: []byte{0xFF}})

}
