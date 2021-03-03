package esbbridgeclient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/spritkopf/esb-bridge/pkg/esbbridge"
)

///////////////////////////////////////////////////////////////////////////////
// Types and constants
///////////////////////////////////////////////////////////////////////////////

// DefaultTcpTimeoutMillis is the default timeout waiting for an answer to a TCP message
const DefaultTcpTimeoutMillis uint32 = 1000

// TCPPackageSizeMax represents the maximum size of a TCP package
const TCPPackageSizeMax uint8 = 64

// TCPPackageSizeMin represents the minimum size of a TCP package
const TCPPackageSizeMin uint8 = 2

const (
	// EsbBridgeCmdTransfer - Transfer ESB message, receive answer
	EsbBridgeCmdTransfer uint8 = 0x10
)

// EsbAddressSize is the size of the Pipeline addresses (only 5 byte addresses are supported)
const EsbAddressSize int = esbbridge.AddressSize

type tcpMessage struct {
	Cmd    byte
	Err    byte
	EsbMsg esbbridge.EsbMessage
}

///////////////////////////////////////////////////////////////////////////////
// Private variables
///////////////////////////////////////////////////////////////////////////////

var ansChannel chan tcpMessage
var connection net.Conn
var connected bool = false
var listeners []esbbridge.Listener // Stores callback channels associated to commandIDs and addresses to listen for

///////////////////////////////////////////////////////////////////////////////
// Public API
///////////////////////////////////////////////////////////////////////////////

// Connect connects to the esb bridge server
// Params:
//   addr: address in form address:port  (e.g. 10.65.188.2:9815)
func Connect(addr string) error {

	conn, err := net.Dial("tcp", addr)

	if err != nil {
		fmt.Printf("Error connecting to %v: %v\n", addr, err)
		return err
	}

	connection = conn
	connected = true

	ansChannel = make(chan tcpMessage)

	go tcpReaderThread(conn, ansChannel)

	return nil
}

// Transfer sends an ESB packet to a target device and waits for the answer
// Params:
//   targetAddr - ESB pipeline address of target device, only 5 bytes address length supported
//   cmd		- command byte for the ESB message
//   payload	- payload of the esb message (can be nil for zero payload, cmd-only message)
func Transfer(targetAddr []byte, cmd byte, payload []byte) ([]byte, error) {
	if !connected {
		return nil, fmt.Errorf("Not connected to server")
	}

	if len(targetAddr) != 5 {
		return nil, fmt.Errorf("Invalid address length (only 5 byte addresses supported)")
	}

	packetBuffer := make([]byte, 2, TCPPackageSizeMax)
	esbPacketBuffer := make([]byte, 6, TCPPackageSizeMax-uint8(len(packetBuffer)))

	copy(esbPacketBuffer[:5], targetAddr)
	esbPacketBuffer[5] = cmd
	esbPacketBuffer = append(esbPacketBuffer, payload...)

	packetBuffer[0] = EsbBridgeCmdTransfer
	packetBuffer[1] = uint8(len(esbPacketBuffer))
	packetBuffer = append(packetBuffer, esbPacketBuffer...)

	connection.Write(packetBuffer)

	// Wait for answer or Timeout
	select {
	case answer := <-ansChannel:
		// check that answer actually matches request (cmdID)
		if answer.Cmd != packetBuffer[0] {
			// Answer command byte must be identical
			return nil, errors.New("Answer message command byte did not match request")
		}
		if answer.Err != 0 {
			return nil, fmt.Errorf("Answer contained error code: %v", answer.Err)
		}

		return answer.EsbMsg.Payload, nil

	case <-time.After(time.Duration(DefaultTcpTimeoutMillis) * time.Millisecond):
		return nil, errors.New("Timeout waiting for an answer")
	}
}

// AddListener adds a listenener. Any incoming message with this CommandID and/or address will be redirected to c
// Params:
//   sourceAddr - only messages from this sender will be evaluated, an empty array is used to ignore this filter (all senders will be evaluated)
//   cmd        - only messages with a specific cmd byte (the 1st payload byte) will be evaluated, set to 0xFF to ignore the filter (all message IDs will be evaluated)
func AddListener(sourceAddr [EsbAddressSize]byte, cmd byte, c esbbridge.ListenerChannel) error {
	connection.Write([]byte{0x20, 6, 12, 13, 14, 15, 16, 0xFF})
	answerBuffer := make([]byte, 3)
	io.ReadFull(connection, answerBuffer)
	fmt.Printf("Answer: %v\n", answerBuffer)

	if c == nil {
		return errors.New("invalid parameter passed for listener channel (nil)")
	}

	listeners = append(listeners, esbbridge.Listener{SourceAddr: sourceAddr, Cmd: cmd, Channel: c})

	return nil

}

// Disconnect closes the connection to the esb bridge server
func Disconnect() {
	if connected {
		connection.Close()
		connected = false
	}
}

//////////////////////////////
// Internal functions (private)
//////////////////////////////

func tcpReaderThread(conn net.Conn, ansChannel chan tcpMessage) {

	for {

		header := make([]byte, 3)
		//_, err := io.ReadAtLeast(conn, header, 2)
		n, err := io.ReadFull(conn, header)

		if err != nil {
			conn.Close()
			return
		}

		if n < 3 {
			log.Printf("Packet too short, need at least 2 bytes (cmd + len), got %v %v, ", n, header)
			continue
		}
		messageCmd := header[0]
		messageError := header[1]
		payloadSize := header[2]

		payload := make([]byte, int(payloadSize))
		if payloadSize > 0 {
			io.ReadFull(conn, payload)
		}

		log.Printf("Incoming message: Cmd %v, payload %v", header[0], payload)

		tcpMsg := tcpMessage{
			Cmd: messageCmd,
			Err: messageError,
			EsbMsg: esbbridge.EsbMessage{
				Address: payload[:5],
				Cmd:     payload[5],
				Payload: payload[6:]}}

		if messageCmd == 0x21 {
			// this is an incoming notification message, check listeners and send the EsbMessage to associated channels
			for _, l := range listeners {
				if ((l.Cmd == 0xFF) || (l.Cmd == tcpMsg.EsbMsg.Cmd)) &&
					((bytes.Compare(l.SourceAddr[:], tcpMsg.EsbMsg.Address) == 0) || (bytes.Compare(l.SourceAddr[:], make([]byte, 5)) == 0)) {
					l.Channel <- tcpMsg.EsbMsg
				}
			}

		} else {
			// this is probably an answer to a command, send to answer channel
			// we are sending the whole tcpMessage here so the recipient can handle possible error codes
			ansChannel <- tcpMsg
		}
	}
}
