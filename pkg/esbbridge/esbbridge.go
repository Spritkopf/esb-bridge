package esbbridge

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/spritkopf/esb-bridge/internal/usbprotocol"
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

// EsbMessage is the data type representing a message sent between esb devices
type EsbMessage struct {
	Address []byte
	Cmd     byte
	Payload []byte
}

// ListenerChannel is used to notify a subscriber about a incoming message it was listening for
type ListenerChannel chan<- EsbMessage // listenerChannel is send-only

// Listener holds all information necessary listening for a specific message
type Listener struct {
	SourceAddr [AddressSize]byte
	Cmd        byte
	Channel    ListenerChannel
}

///////////////////////////////////////////////////////////////////////////////
// Private variables
///////////////////////////////////////////////////////////////////////////////

var connected bool = false
var listeners []Listener // Stores callback channels associated to commandIDs and addresses to listen for

///////////////////////////////////////////////////////////////////////////////
// Public API
///////////////////////////////////////////////////////////////////////////////

// Open opens the connection to the esb bridge device
// Parameters:
//   device	- device string , e.g. "/dev/ttyACM0"
func Open(device string) error {
	err := usbprotocol.Open(device)

	if err != nil {
		return fmt.Errorf("Could not connect to device %v: %v", device, err)
	}
	connected = true

	rxChannel := make(chan usbprotocol.Message, 5)
	// start listening for all incoming messages with Command ID "CmdRx"
	err = usbprotocol.AddListener(usbprotocol.CmdRx, rxChannel)

	go rxCallbackThread(rxChannel)

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

	txMsg := usbprotocol.Message{}
	txMsg.Cmd = UsbCmdVersion
	answerMessage, err := usbprotocol.Transfer(txMsg)

	if answerMessage.Err != 0x00 {
		return "", fmt.Errorf("Command CmdVersion (0x%02X) returned Error 0x%02X", UsbCmdVersion, answerMessage.Err)
	}

	if err != nil {
		return "", err
	}
	versionStr := fmt.Sprintf("%v.%v.%v", answerMessage.Payload[0], answerMessage.Payload[1], answerMessage.Payload[2])
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

	txMsg := usbprotocol.Message{}
	txMsg.Cmd = UsbCmdTransfer
	txMsg.Payload = append(txMsg.Payload, targetAddr[:]...)
	txMsg.Payload = append(txMsg.Payload, payload[:]...)

	answerMessage, err := usbprotocol.Transfer(txMsg)

	if answerMessage.Err != 0 {
		return nil, fmt.Errorf("ESB Transfer command returned with error code: 0x%02X", answerMessage.Err)
	}

	if err != nil {
		return nil, err
	}

	return answerMessage.Payload, nil
}

// AddListener adds a listenener. Any incoming message with this CommandID and/or address will be redirected to c
// Params:
//   sourceAddr - only messages from this sender will be evaluated, an empty array is used to ignore this filter (all senders will be evaluated)
//   cmd        - only messages with a specific cmd byte (the 1st payload byte) will be evaluated, set to 0xFF to ignore the filter (all message IDs will be evaluated)
func AddListener(sourceAddr [AddressSize]byte, cmd byte, c ListenerChannel) error {

	if c == nil {
		return errors.New("invalid parameter passed for listener channel (nil)")
	}

	listeners = append(listeners, Listener{SourceAddr: sourceAddr, Cmd: cmd, Channel: c})

	return nil
}

// RemoveListener removes a listenener. Any listener which was registered for the specified channel will be deleted.
// Returns the number of deleted listeners
func RemoveListener(c ListenerChannel) int {

	var itemsDeleted int = 0
searchLoop:
	for {
		for i, l := range listeners {
			if l.Channel == c {
				// listener channel matches, remove item
				listeners = append(listeners[:i], listeners[i+1:]...)
				itemsDeleted++
				// restart search since the listeners slice is shorter now
				continue searchLoop
			}
		}
		// if no more occurences are found, we are finished
		break searchLoop
	}

	return itemsDeleted
}

///////////////////////////////////////////////////////////////////////////////
// Private functions
///////////////////////////////////////////////////////////////////////////////

func rxCallbackThread(ch chan usbprotocol.Message) {

	for {
		usbMsg := <-ch

		// message error is discarded for CmdRx, should always be OK

		// check payload size, must at least contain a source address (5 bytes) and a cmd ID
		if len(usbMsg.Payload) < 6 {
			return
		}

		message := EsbMessage{}

		message.Address = usbMsg.Payload[:5]
		message.Cmd = usbMsg.Payload[5]

		if len(usbMsg.Payload) > 6 {
			message.Payload = usbMsg.Payload[6:]
		}

		// send message to all registered and matching listeners
		for _, l := range listeners {
			if ((l.Cmd == 0xFF) || (l.Cmd == message.Cmd)) &&
				((bytes.Compare(l.SourceAddr[:], message.Address) == 0) || (bytes.Compare(l.SourceAddr[:], make([]byte, 5)) == 0)) {
				l.Channel <- message
			}
		}
	}
}
