package esbbridge

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/spritkopf/esb-bridge/usbprotocol"
)

///////////////////////////////////////////////////////////////////////////////
// Types and constants
///////////////////////////////////////////////////////////////////////////////

// AddressSize is the size of the Pipeline addresses (only 5 byte addresses are supported)
const AddressSize int = 5

// MaxPayloadSize represents the maximum amount of bytes which fit into the payload of an ESB message
// The maximum payload size is limited to 32 by the implementation of the ESB protocol on the nRF52 uC
const MaxPayloadSize uint8 = 32

const (
	// UsbCmdVersion - Get firmware version
	UsbCmdVersion usbprotocol.CommandID = 0x10
	// UsbCmdTransfer - Send a message, wait for reply
	UsbCmdTransfer usbprotocol.CommandID = 0x30
	// UsbCmdSend - Send a message without reply
	UsbCmdSend usbprotocol.CommandID = 0x31
	// UsbCmdRx - callback from incoming ESB message
	UsbCmdRx usbprotocol.CommandID = 0x81
)

// EsbRxMessageCallback - function prototype for incoming message callbacks
// When called the function gets passed the error byte of the origin of the message (source address), command byte, and the payload
type EsbRxMessageCallback func(EsbMessage)

// the callback type is used by the receive routine to map command IDs to callback functions
type callback struct {
	sourceAddr [AddressSize]byte
	cmd        byte
	cbFunc     EsbRxMessageCallback
}

// EsbMessage is the data type representing a message sent between esb devices
type EsbMessage struct {
	address []byte
	cmd     byte
	payload []byte
}

///////////////////////////////////////////////////////////////////////////////
// Private variables
///////////////////////////////////////////////////////////////////////////////

var connected bool = false
var callbacks []callback

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

	// start listening for all incoming messages with Command ID "CmdRX"
	err = usbprotocol.RegisterCallback(usbprotocol.CmdRx, rxCallback)

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

	answerErr, answerPayload, err := usbprotocol.Transfer(UsbCmdVersion, nil)

	if answerErr != 0x00 {
		return "", fmt.Errorf("Command CmdVersion (0x%02X) returned Error 0x%02X", UsbCmdVersion, answerErr)
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

	if payload == nil {
		return nil, fmt.Errorf("Payload must not be empty")
	}

	if len(payload) > int(MaxPayloadSize) {
		return nil, fmt.Errorf("Payload too long, maximum is %v", MaxPayloadSize)
	}
	if len(payload) < 1 {
		return nil, fmt.Errorf("Payload too short, minimum is 6 (5bytes address, at least 1 byte payload)")
	}

	txMsg := make([]byte, AddressSize+len(payload))
	copy(txMsg[0:5], targetAddr[:])
	copy(txMsg[5:], payload[:])
	answerErr, answerPayload, err := usbprotocol.Transfer(UsbCmdTransfer, txMsg)

	if answerErr != 0 {
		return nil, fmt.Errorf("ESB Transfer command returned with error code: 0x%02X", answerErr)
	}

	if err != nil {
		return nil, err
	}

	return answerPayload, nil
}

// RegisterCallback registers a callback function to call when a specific message arrives
// Params:
//   sourceAddr - only messages from this sender will be evaluated, an empty array is used to ignore this filter (all senders will be evaluated)
//   cmd        - only messages with a specific cmd byte (the 1st payload byte) will be evaluated, set to 0xFF to ignore the filter (all message IDs will be evaluated)
func RegisterCallback(sourceAddr [AddressSize]byte, cmd byte, callbackFunc EsbRxMessageCallback) error {

	if callbackFunc == nil {
		return errors.New("invalid parameter passed for callbackFunc")
	}
	callbacks = append(callbacks, callback{sourceAddr, cmd, callbackFunc})

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// Private functions
///////////////////////////////////////////////////////////////////////////////

func rxCallback(msgErr byte, usbPayload []byte) {

	// message error is discarded for CmdRx, should always be OK

	// check payload size, must at least contain a source address and a cmd ID
	if len(usbPayload) < 6 {
		return
	}

	message := EsbMessage{}

	message.address = usbPayload[:5]
	message.cmd = usbPayload[5]

	if len(usbPayload) > 6 {
		message.payload = usbPayload[6:]
	}

	// go through all registered callbacks, check if the cmd ID matches (or is ignored) and the source address matches (or is ignored)
	for _, cb := range callbacks {
		if ((cb.cmd == 0xFF) || (cb.cmd == message.cmd)) &&
			((bytes.Compare(cb.sourceAddr[:], message.address) == 0) || (bytes.Compare(cb.sourceAddr[:], make([]byte, 5)) == 0)) {
			cb.cbFunc(message)
		}
	}
}
