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
	Error   byte
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

func (m EsbMessage) String() string {
	return fmt.Sprintf("Addr: %v Cmd: %v, Error: %v, Payload: %v", m.Address, m.Cmd, m.Error, m.Payload)
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

// Transfer sends a message to an ESB device and returns the answer
func Transfer(message EsbMessage) (EsbMessage, error) {
	if !connected {
		return EsbMessage{}, errors.New("Device is not connected, call Open() first")
	}

	if len(message.Payload) > int(MaxPayloadSize) {
		return EsbMessage{}, fmt.Errorf("Payload too long, maximum is %v", MaxPayloadSize)
	}

	if message.Payload == nil {
		message.Payload = []byte{}
	}

	txMsg := usbprotocol.Message{}
	txMsg.Cmd = UsbCmdTransfer
	txMsg.Payload = append(txMsg.Payload, message.Address...)
	txMsg.Payload = append(txMsg.Payload, message.Cmd)
	txMsg.Payload = append(txMsg.Payload, message.Payload...)

	answerMessage, err := usbprotocol.Transfer(txMsg)

	if err != nil {
		return EsbMessage{}, err
	}

	if answerMessage.Err != 0 {
		return EsbMessage{}, fmt.Errorf("ESB Transfer command returned with error code: 0x%02X", answerMessage.Err)
	}

	message.Cmd = answerMessage.Payload[0]
	message.Error = answerMessage.Payload[1]
	message.Payload = answerMessage.Payload[2:]
	return message, nil
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

		// check payload size, must at least contain a source address (5 bytes), error, and a cmd ID
		if len(usbMsg.Payload) < 7 {
			return
		}

		message := EsbMessage{}

		message.Cmd = usbMsg.Payload[0]
		// message error (usbMsg.Payload[1]) is discarded for CmdRx, should always be OK
		message.Address = usbMsg.Payload[2:7]

		if len(usbMsg.Payload) > 7 {
			message.Payload = usbMsg.Payload[7:]
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
