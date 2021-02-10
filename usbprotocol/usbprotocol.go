package usbprotocol

import (
	"errors"
	"fmt"

	"github.com/sigurn/crc16"
	"go.bug.st/serial.v1"
)

// packetSize is the fixed size of transmitted USB packages
const packetSize = 64

// MaxPayloadLen - maximum length of message Payload (64 byte packet - 4 bytes header - 2 bytes crc)
const MaxPayloadLen = packetSize - 4 - 2

// sync byte, marks the beginning of a new packet
const sync = 0x69

////////////////////////////
// Type definitions
///////////////////////////

// SizeError is returned, when input data is of invalid size (e.g. message payload for transfer)
type SizeError uint32

func (f SizeError) Error() string {
	return fmt.Sprintf("Size of passed data too large, allowed %v, got %v", MaxPayloadLen, uint32(f))
}

// CommandID - ID of the USB commands
type CommandID uint8

const (
	// CmdVersion - Get firmware version
	CmdVersion CommandID = 0x10
	// CmdTransfer - Send a message, wait for reply
	CmdTransfer CommandID = 0x30
	// CmdSend - Send a message without reply
	CmdSend CommandID = 0x31
	// CmdTest - test command, do not use
	CmdTest CommandID = 0x61
	// CmdIrq - interrupt callback, only from device to host
	CmdIrq CommandID = 0x80
	// CmdRx - callback from incoming ESB message
	CmdRx CommandID = 0x81
)

// IncomingMessageCallback - function prototype for incoming message callbacks
// When called the function gets passed the error byte of the message and the payload
type IncomingMessageCallback func(err byte, payload []byte)

/////////////////////////////
// Package variables (private)
/////////////////////////////
var crcTable *crc16.Table
var port serial.Port

var testCallback IncomingMessageCallback

/////////////////////////////
// Package API (public)
/////////////////////////////

// Open connects to the specified virtual COM port
// The parameter 'device' holds the name of the device to connect to, i.e. '/dev/ttyACM0'
func Open(device string) error {
	var err error
	// Open port in mode 115200_N81
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err = serial.Open(device, mode)

	return err
}

// Close closes the connection to any opened virtual COM port
func Close() {
	port.Close()
}

// Transfer sends a message to the usb device and returns the answer
// Returnvalues are Answer-ErrorCode, Payload, error
func Transfer(cmd CommandID, payload []byte) (byte, []byte, error) {
	if len(payload) > MaxPayloadLen {
		return 0, nil, SizeError(len(payload))
	}

	var txBuf [packetSize]byte
	txBuf[0] = sync
	txBuf[1] = byte(cmd)
	txBuf[2] = 0
	txBuf[3] = byte(len(payload))
	copy(txBuf[4:], payload[:])
	crc := crc16.Checksum(txBuf[:len(txBuf)-2], crcTable)
	var h, l uint8 = uint8(crc & 0xff), uint8(crc >> 8)
	txBuf[62] = byte(h)
	txBuf[63] = byte(l)

	bytesWritten, err := port.Write(txBuf[:])

	// Send the message
	if err != nil || bytesWritten != len(txBuf) {
		return 0, nil, err
	}

	// Receive answer
	var rxBuf [packetSize]byte
	bytesRead, err := port.Read(rxBuf[:])
	if err != nil || bytesRead != len(rxBuf) {
		return 0, nil, err
	}

	// Check answer for errors
	if rxBuf[1] != txBuf[1] {
		// Answer command byte must be identical
		return 0, nil, err
	}
	answerLen := rxBuf[3]
	return rxBuf[2], rxBuf[4 : 4+answerLen], nil
}

func waitForMessage() {
	// Receive answer
	var rxBuf [packetSize]byte

	bytesRead, err := port.Read(rxBuf[:])

	if err != nil || bytesRead != len(rxBuf) {
		return
	}
	answerLen := rxBuf[3]
	if testCallback != nil {
		testCallback(rxBuf[1], rxBuf[4:4+answerLen])
	}
}

// RegisterCallback registers a function which is called when message with a certain CommandId is incoming
func RegisterCallback(cmd CommandID, callback IncomingMessageCallback) error {

	if callback == nil {
		return errors.New("Callback parameter should not be nil")
	}

	testCallback = callback

	// temporary start reader thread here, should be done in Open()
	go waitForMessage()

	return nil
}

//////////////////////////////
// Internal functions (private)
//////////////////////////////

func init() {
	// create crc16 table
	crcTable = crc16.MakeTable(crc16.CRC16_CCITT_FALSE)
}
