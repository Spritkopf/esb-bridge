// Package client implements a simple gRPC client that implements the esbbridge rpc client described in esbbridge_rpc.proto
package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/spritkopf/esb-bridge/pkg/esbbridge"
	pb "github.com/spritkopf/esb-bridge/pkg/server/service"
)

// Transfer sends a message to a peripheral device and returns the answer message
func Transfer(client pb.EsbBridgeClient, msg *pb.EsbMessage) (esbbridge.EsbMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	answerMessage, err := client.Transfer(ctx, msg)
	if err != nil {
		log.Fatalf("%v.Transfer(_) = _, %v: ", client, err)
		return esbbridge.EsbMessage{}, err
	}
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
