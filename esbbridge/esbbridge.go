package esbbridge

import (
	"errors"
	"fmt"

	"github.com/spritkopf/esb-bridge-client/usbprotocol"
)

///////////////////////////////////////////////////////////////////////////////
// Types and constants
///////////////////////////////////////////////////////////////////////////////

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

///////////////////////////////////////////////////////////////////////////////
// Private functions
///////////////////////////////////////////////////////////////////////////////
