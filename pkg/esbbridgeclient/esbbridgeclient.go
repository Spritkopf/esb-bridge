package esbbridgeclient

import (
	"fmt"
	"net"
)

///////////////////////////////////////////////////////////////////////////////
// Types and constants
///////////////////////////////////////////////////////////////////////////////

// TCPPackageSizeMax represents the maximum size of a TCP package
const TCPPackageSizeMax uint8 = 64

// TCPPackageSizeMin represents the minimum size of a TCP package
const TCPPackageSizeMin uint8 = 2

const (
	// EsbBridgeCmdTransfer - Transfer ESB message, receive answer
	EsbBridgeCmdTransfer uint8 = 0x00
)

///////////////////////////////////////////////////////////////////////////////
// Private variables
///////////////////////////////////////////////////////////////////////////////

var connection net.Conn
var connected bool = false

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

	return nil
}

// Transfer sends an ESB packet to a target device and waits for the answer
// Params:
//   targetAddr - ESB pipeline address of target device, only 5 bytes address length supported
//   cmd		- command byte for the ESB message
//   payload	- payload of the esb message
func Transfer(targetAddr []byte, cmd byte, payload []byte) error {
	if !connected {
		return fmt.Errorf("Not connected to server")
	}

	if len(targetAddr) != 5 {
		return fmt.Errorf("Invalid address length (only 5 byte addresses supported)")
	}

	packetBuffer := make([]byte, 2, TCPPackageSizeMax)
	esbPacketBuffer := make([]byte, 7, TCPPackageSizeMax-uint8(len(packetBuffer)))

	copy(esbPacketBuffer[:5], targetAddr)
	esbPacketBuffer[5] = cmd
	esbPacketBuffer[6] = uint8(len(payload))
	esbPacketBuffer = append(esbPacketBuffer, payload...)

	packetBuffer[0] = 0x00
	packetBuffer[1] = uint8(len(esbPacketBuffer))
	packetBuffer = append(packetBuffer, esbPacketBuffer...)

	connection.Write(packetBuffer)

	return nil

}

// Disconnect closes the connection to the esb bridge server
func Disconnect() {
	if connected {
		connection.Close()
		connected = false
	}
}
