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
	"google.golang.org/grpc"
)

//////////////////////////////////////////////////////////
// Public members
//////////////////////////////////////////////////////////

// DefaultTimeout is the timeout used for RPC activities like connection, or transfers
var DefaultTimeout time.Duration = 2 * time.Second

//////////////////////////////////////////////////////////
// Private variables
//////////////////////////////////////////////////////////
var conn *grpc.ClientConn
var client pb.EsbBridgeClient
var connected bool = false

// Connect establishes a connection to the ESB bridge RPC server.
// This function must be called first in order to use this package
// Param:
//   address: remote address and port of the server, e.g. "localhost:10000"
func Connect(address string) error {
	var err error
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(DefaultTimeout))

	conn, err = grpc.Dial(address, opts...)
	if err != nil {
		return fmt.Errorf("Could not connect to esb-bridge RPC server: %v", err)
	}
	client = pb.NewEsbBridgeClient(conn)

	connected = true

	return nil
}

// Disconnect terminates the TCP connection to the server
func Disconnect() error {

	if !connected {
		return fmt.Errorf("Not connected to server")
	}
	err := conn.Close()
	if err != nil {
		return fmt.Errorf("Error while disconnection: %v)", err)
	}

	connected = false

	return nil
}

// Transfer sends a message to a peripheral device and returns the answer message
func Transfer(client pb.EsbBridgeClient, msg *pb.EsbMessage) (esbbridge.EsbMessage, error) {

	if !connected {
		return esbbridge.EsbMessage{}, fmt.Errorf("Not connected to server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	answerMessage, err := client.Transfer(ctx, msg)
	if err != nil {
		log.Fatalf("%v.Transfer(_) = _, %v: ", client, err)
		return esbbridge.EsbMessage{}, err
	}

	return esbbridge.EsbMessage{Address: answerMessage.Addr, Cmd: answerMessage.Cmd[0], Error: answerMessage.Error[0], Payload: answerMessage.Payload}, nil
}

// Listen will start a listening goroutine which listens for specific messages and sends them to the channel returned by Listen().
// The RPC Message stream will keep running indefinitely until the context is cancelled. Use context.WithCancel and call the cancelFunc.
// When the context is cancelled, the RPC stream is terminated and the server will stop listening for these messages
func Listen(ctx context.Context, client pb.EsbBridgeClient, listener *pb.Listener) (<-chan esbbridge.EsbMessage, error) {

	if !connected {
		return nil, fmt.Errorf("Not connected to server")
	}
	stream, err := client.Listen(ctx, listener)
	if err != nil {
		fmt.Printf("%v.Listen(_) = _, %v", client, err)
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
				fmt.Printf("%v.ListenWorker(_) = _, %v", client, err)
				return
			}
			fmt.Printf("Incoming Message: %v", incomingMessage)
			answerMessage := esbbridge.EsbMessage{
				Address: incomingMessage.Addr,
				Cmd:     incomingMessage.Cmd[0],
				Payload: incomingMessage.Payload}
			rxChan <- answerMessage
		}
	}()

	return rxChan, nil
}
