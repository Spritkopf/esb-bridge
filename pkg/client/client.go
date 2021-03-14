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
// Types and interfaces
//////////////////////////////////////////////////////////

// EsbClientInterface defines the interface functions
type EsbClientInterface interface {
	Connect(address string) error
	Disconnect() error
	Transfer(msg esbbridge.EsbMessage) (esbbridge.EsbMessage, error)
	Listen(ctx context.Context, addr []byte, cmd byte) (<-chan esbbridge.EsbMessage, error)
}

// EsbClient represents the RPC connection and implements the EsbClientInterface
type EsbClient struct {
	conn      *grpc.ClientConn
	client    pb.EsbBridgeClient
	connected bool
}

//////////////////////////////////////////////////////////
// Public members
//////////////////////////////////////////////////////////

// DefaultTimeout is the timeout used for RPC activities like connection, or transfers
var DefaultTimeout time.Duration = 2 * time.Second

// Connect establishes a connection to the ESB bridge RPC server.
// This function must be called first in order to use this package
// Param:
//   address: remote address and port of the server, e.g. "localhost:10000"
func (c *EsbClient) Connect(address string) error {
	var err error
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(DefaultTimeout))

	c.conn, err = grpc.Dial(address, opts...)
	if err != nil {
		return fmt.Errorf("Could not connect to esb-bridge RPC server: %v", err)
	}
	c.client = pb.NewEsbBridgeClient(c.conn)

	c.connected = true

	return nil
}

// Disconnect terminates the TCP connection to the server
func (c *EsbClient) Disconnect() error {

	if !c.connected {
		return fmt.Errorf("Not connected to server")
	}
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("Error while disconnection: %v)", err)
	}

	c.connected = false

	return nil
}

// Transfer sends a message to a peripheral device and returns the answer message
func (c *EsbClient) Transfer(msg esbbridge.EsbMessage) (esbbridge.EsbMessage, error) {

	if !c.connected {
		return esbbridge.EsbMessage{}, fmt.Errorf("Not connected to server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	answerMessage, err := c.client.Transfer(ctx, &pb.EsbMessage{Addr: msg.Address, Cmd: []byte{msg.Cmd}, Payload: msg.Payload})
	if err != nil {
		log.Fatalf("%v.Transfer(_) = _, %v: ", c.client, err)
		return esbbridge.EsbMessage{}, err
	}

	return esbbridge.EsbMessage{Address: answerMessage.Addr, Cmd: answerMessage.Cmd[0], Error: answerMessage.Error[0], Payload: answerMessage.Payload}, nil
}

// Listen will start a listening goroutine which listens for specific messages and sends them to the channel returned by Listen().
// The RPC Message stream will keep running indefinitely until the context is cancelled. Use context.WithCancel and call the cancelFunc.
// When the context is cancelled, the RPC stream is terminated and the server will stop listening for these messages
func (c *EsbClient) Listen(ctx context.Context, addr []byte, cmd byte) (<-chan esbbridge.EsbMessage, error) {

	if !c.connected {
		return nil, fmt.Errorf("Not connected to server")
	}

	stream, err := c.client.Listen(ctx, &pb.Listener{Addr: addr, Cmd: []byte{cmd}})
	if err != nil {
		fmt.Printf("%v.Listen(_) = _, %v", c.client, err)
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
				fmt.Printf("%v.ListenWorker(_) = _, %v", c.client, err)
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
