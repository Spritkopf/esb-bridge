package esbbridgeclient

import (
	"fmt"
	"net"
)

///////////////////////////////////////////////////////////////////////////////
// Types and constants
///////////////////////////////////////////////////////////////////////////////

// TcpPackageSize represents the maximum size of a TCP package
const TcpPackageSizeMax uint8 = 64

// TcpPackageSizeMin represents the minimum size of a TCP package
const TcpPackageSizeMin uint8 = 2

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

// Disconnect closes the connection to the esb bridge server
func Disconnect() {
	if connected {
		connection.Close()
		connected = false
	}
}
