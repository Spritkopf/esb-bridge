package esbbridge

import (
	"errors"
	"fmt"

	"github.com/spritkopf/esb-bridge-client/usbprotocol"
)

///////////////////////////////////////////////////////////////////////////////
// Types and constants
///////////////////////////////////////////////////////////////////////////////

// AddressSize is the size of the Pipeline addresses (only 5 byte addresses are supported)
const AddressSize uint8 = 5

// MaxPayloadSize represents the maximum amount of bytes which fit into the payload of an ESB message
// The maximum payload size is limited to 32 by the implementation of the ESB protocol on the nRF52 uC
const MaxPayloadSize uint8 = 32

const (
	// CmdVersion - Get firmware version
	CmdVersion usbprotocol.CommandID = 0x10
	// CmdTransfer - Send a message, wait for reply
	CmdTransfer usbprotocol.CommandID = 0x30
	// CmdSend - Send a message without reply
	CmdSend usbprotocol.CommandID = 0x31
	// CmdRx - callback from incoming ESB message
	CmdRx usbprotocol.CommandID = 0x81
)

///////////////////////////////////////////////////////////////////////////////
// Private variables
///////////////////////////////////////////////////////////////////////////////

var connected bool = false

///////////////////////////////////////////////////////////////////////////////
// Public API
///////////////////////////////////////////////////////////////////////////////

// Open opens the connection to the esb bridge device
// Parameters:
//   device	- device string , e.g. "/dev/ttyACM0"
func Open(device string) error {
	err := usbprotocol.Open(device)

	if err == nil {
		connected = true
	}
	return err
}

// Close closes the connection to the esb bridge device
func Close() {
	usbprotocol.Close()
}

// GetFwVersion reads the firmware version of the conected esb-bridge
// Returns the firmware version as string in format "maj.min.patch"
func GetFwVersion() (string, error) {
	if !connected {
		return "", errors.New("Device is not connected, call Open() first")
	}

	answerErr, answerPayload, err := usbprotocol.Transfer(CmdVersion, nil)

	if answerErr != 0x00 {
		return "", fmt.Errorf("Command CmdVersion (0x%02X) returned Error 0x%02X", CmdVersion, answerErr)
	}

	if err != nil {
		return "", err
	}
	versionStr := fmt.Sprintf("%v.%v.%v", answerPayload[0], answerPayload[1], answerPayload[2])
	return versionStr, nil
}

// Transfer sends a packet to the target pipeline address and returns the answer
//
// Params:
//   targetAddr - target pipeline address, only 5-byte addresses are supported
//   payload    - payload to be transmitted, maximum length is 32 (see MaxPayloadSize)
// Returns a slice of bytes with the answer payload and an error
func Transfer(targetAddr [AddressSize]byte, payload []byte) ([]byte, error) {
	if !connected {
		return nil, errors.New("Device is not connected, call Open() first")
	}

	if len(payload) > int(MaxPayloadSize) {
		return nil, fmt.Errorf("Payload too long, maximum is %v", MaxPayloadSize)
	}

	answerErr, answerPayload, err := usbprotocol.Transfer(CmdTransfer, payload)

	if answerErr != 0 {
		return nil, fmt.Errorf("ESB Transfer command returned with error code: %02X", answerErr)
	}

	if err != nil {
		return nil, err
	}

	return answerPayload, nil
}

///////////////////////////////////////////////////////////////////////////////
// Private functions
///////////////////////////////////////////////////////////////////////////////
