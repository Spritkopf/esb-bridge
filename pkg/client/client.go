// Package main implements a simple gRPC client that implements the esbbridge rpc client described in esbbridge_rpc.proto
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/spritkopf/esb-bridge/pkg/esbbridge"
	pb "github.com/spritkopf/esb-bridge/pkg/server/service"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
)

// transfer sends a message to a peripheral device and returns the answer message
func transfer(client pb.EsbBridgeClient, msg *pb.EsbMessage) (esbbridge.EsbMessage, error) {
	log.Printf("Sending Message %v", msg)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	answerMessage, err := client.Transfer(ctx, msg)
	if err != nil {
		log.Fatalf("%v.Transfer(_) = _, %v: ", client, err)
		return esbbridge.EsbMessage{}, err
	}
	log.Printf("Answer: %v\n", answerMessage)
	return esbbridge.EsbMessage{Address: answerMessage.Addr, Cmd: answerMessage.Cmd[0], Payload: answerMessage.Payload}, nil
}

// Listen will start a listening goroutine which listens for specific messages and sends them to the channel returned by Listen().
// The RPC Message stream will kep running indefinitely until the context is canceled. Use context.WithCancel and call the cancelFunc.
// When the context is cancelled, the ROX stream is terminated and the server will stop listening for these messages
func Listen(ctx context.Context, client pb.EsbBridgeClient, listener *pb.Listener) (<-chan esbbridge.EsbMessage, error) {
	log.Printf("Start listening: %v", listener)
	stream, err := client.Listen(ctx, listener)
	if err != nil {
		log.Printf("%v.Listen(_) = _, %v", client, err)
		return nil, fmt.Errorf("Error calling remote procedure `Listen()`: %v", err)
	}

	rxChan := make(chan esbbridge.EsbMessage, 1)

	go func() {
		for {
			incomingMessage, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatalf("%v.ListenWorker(_) = _, %v", client, err)
			}
			log.Printf("Incoming Message: %v", incomingMessage)
			answerMessage := esbbridge.EsbMessage{
				Address: incomingMessage.Addr,
				Cmd:     incomingMessage.Cmd[0],
				Payload: incomingMessage.Payload}
			rxChan <- answerMessage
		}
	}()

	return rxChan, nil
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

	// test Listen function
	ctx, cancel := context.WithCancel(context.Background())
	rxChan, _ := Listen(ctx, client, &pb.Listener{Addr: []byte{12, 13, 14, 15, 16}, Cmd: []byte{0xFF}})

	for i := 0; i < 4; i++ {
		msg := <-rxChan
		log.Printf("Incoming Message: %v", msg)
	}
	cancel()

}
