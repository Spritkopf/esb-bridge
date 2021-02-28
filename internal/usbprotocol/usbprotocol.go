package usbprotocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/sigurn/crc16"

	"github.com/tarm/serial"
)

// packetSize is the fixed size of transmitted USB packages
const packetSize = 64

// MaxPayloadLen - maximum length of message Payload (64 byte packet - 4 bytes header - 2 bytes crc)
const MaxPayloadLen = packetSize - 4 - 2

// DefaultTimeout is the default Transfer-timeout in milliseconds waiting for an answer message befor returning an error
const DefaultTimeout = 500

// sync byte, marks the beginning of a new packet
const sync = 0x69

const idxSync = 0
const idxCmd = 1
const idxErr = 2
const idxlen = 3
const idxPayload = 4

////////////////////////////
// Type definitions
///////////////////////////

// ErrSize usually is returned when the passed data has the wrong size (too large or too small)
var ErrSize = UsbError{1, errors.New("ErrSize: Invalid size")}

// ErrCmdMismatch is returned when the command ID of a received answer doesn't match the request
var ErrCmdMismatch = UsbError{2, errors.New("ErrCmdMismatch: Unexpected answer command ID")}

// ErrSerial is returned when there is a problem with the serial port
var ErrSerial = UsbError{3, errors.New("ErrSerial: Error while accessing serial port")}

// ErrTimeout is returned when waiting for an answer timed out
var ErrTimeout = UsbError{4, errors.New("ErrTimeout: Timeout while waiting for answer")}

// ErrParam is returned when a passed parameter is invalid
var ErrParam = UsbError{5, errors.New("ErrParam: Invalid Parameter")}

// UsbError is the general Error type for this package.
// Member ErrCode is the specific error code to tell them apart
type UsbError struct {
	ErrCode int
	Err     error
}

func (e UsbError) Error() string {
	return fmt.Sprintf("%v (%d)", e.Err.Error(), e.ErrCode)
}

// CommandID - ID of the USB commands
type CommandID uint8

const (
	// CmdTest - test command, do not use, used for tests
	CmdTest CommandID = 0x61
	// CmdIrq - interrupt callback, only from device to host, used for tests
	CmdIrq CommandID = 0x80
	// CmdRx - Rx callback, for async messages from peripheral -> central
	CmdRx CommandID = 0x81
)

// Message represents a message which is sent between host and device
type Message struct {
	Cmd     CommandID
	Err     uint8
	Payload []byte
}

type listenerChannel chan<- Message

type listener struct {
	cmd      CommandID
	channels []listenerChannel
}

/////////////////////////////
// Package variables (private)
/////////////////////////////
var crcTable *crc16.Table
var port *serial.Port

var rxChannel chan Message  // Used to pass incoming serial messages from the readerThread to the receive goroutine
var ansChannel chan Message // Used to pass incoming serial messages as answer from the the receive goroutine to the transfer function
var listeners []listener    // Stores callback channels associated to command IDs to listen for

/////////////////////////////
// Package API (public)
/////////////////////////////

// TimeoutMillis is the timeout in milliseconds used when waiting for an answer in Transfer()
var TimeoutMillis uint32 = DefaultTimeout

// Open connects to the specified virtual COM port
// The parameter 'device' holds the name of the device to connect to, i.e. '/dev/ttyACM0'
func Open(device string) error {
	var err error
	// Open port in mode 115200_N81
	c := &serial.Config{Name: device, Baud: 115200, ReadTimeout: time.Millisecond * 500}
	port, err = serial.OpenPort(c)

	if err == nil {
		// Start reader goroutine, which sends incoming messages on rxChannel
		rxChannel = make(chan Message)
		ansChannel = make(chan Message)

		go serialReaderThread()
	}

	return err
}

// Close closes the connection to any opened virtual COM port
func Close() {
	if port != nil {
		port.Close()
	}
}

// Transfer sends a message to the usb device and returns the answer
//
// Params:
//   msg - The messge to be transmitted (payload can be nil for zero TX payload (request-only style commands))
// Returns: answer message, error
func Transfer(msg Message) (Message, error) {
	if len(msg.Payload) > MaxPayloadLen {
		return Message{}, ErrSize
	}
	txBuf := make([]byte, packetSize)

	txBuf[0] = sync
	txBuf[1] = byte(msg.Cmd)
	txBuf[2] = 0

	if msg.Payload == nil {
		txBuf[3] = 0
	} else {
		txBuf[3] = byte(len(msg.Payload))
		copy(txBuf[4:], msg.Payload[:])
	}

	crc := crc16.Checksum(txBuf[:len(txBuf)-2], crcTable)
	var h, l uint8 = uint8(crc & 0xff), uint8(crc >> 8)
	txBuf[62] = byte(h)
	txBuf[63] = byte(l)

	// Send the message
	bytesWritten, err := port.Write(txBuf)

	if err != nil {
		return Message{}, err
	}

	if bytesWritten != len(txBuf) {
		return Message{}, ErrSerial
	}

	// Wait for answer or Timeout
	select {
	case answer := <-ansChannel:
		// check that answer actually matches request (cmdID)
		if answer.Cmd != msg.Cmd {
			// Answer command byte must be identical
			return Message{}, ErrCmdMismatch
		}
		return answer, nil

	case <-time.After(time.Duration(TimeoutMillis) * time.Millisecond):
		// timeout, flush port
		return Message{}, ErrTimeout
	}

}

// AddListener adds a listenener for the provided command. Any incoming message with this CommandID will
// sent to the provided channel
func AddListener(cmd CommandID, c listenerChannel) error {

	// If a listener for this command was already registered, just add the channel to it
	for i, l := range listeners {
		if l.cmd == cmd {
			listeners[i].channels = append(listeners[i].channels, c)
			return nil
		}
	}

	// If no listener for this command is already registered, create it
	l := listener{cmd: cmd, channels: []listenerChannel{c}}
	listeners = append(listeners, l)
	return nil
}

//////////////////////////////
// Internal functions (private)
//////////////////////////////

func serialReaderThread() {

	for {
		var rxBuf [packetSize]byte

		if port != nil {
			bytesRead, err := port.Read(rxBuf[:])
			//bytesRead, err := io.ReadAtLeast(port, rxBuf[:], 10)
			//_, err := io.ReadAtLeast(conn, header, 2)
			// check packet length, must be 64
			if err != nil || bytesRead != packetSize {
				continue
			}

			// check sync byte
			if rxBuf[idxSync] != sync {
				continue
			}

			// check CRC
			crcCalc := crc16.Checksum(rxBuf[:packetSize-2], crcTable)
			crcRx := binary.LittleEndian.Uint16(rxBuf[packetSize-2:])
			if crcCalc != crcRx {
				continue
			}

			// Get payload length
			payloadLen := rxBuf[3]

			answerMessage := Message{
				Cmd:     CommandID(rxBuf[idxCmd]),
				Err:     rxBuf[idxErr],
				Payload: rxBuf[idxPayload : idxPayload+payloadLen]}

			isAnswer := true
			// message received, look if a listener is registered
			for _, l := range listeners {
				if l.cmd == answerMessage.Cmd {
					// listener found, send message to all associated channels
					for _, c := range l.channels {
						c <- answerMessage
					}
					isAnswer = false
				}
			}
			if isAnswer {
				ansChannel <- answerMessage
			}
		}

	}
}

func init() {
	// create crc16 table
	crcTable = crc16.MakeTable(crc16.CRC16_CCITT_FALSE)
}
