package usbprotocol

import (
	"errors"
	"fmt"

	"encoding/binary"

	"github.com/sigurn/crc16"

	"github.com/tarm/serial"
)

// packetSize is the fixed size of transmitted USB packages
const packetSize = 64

// MaxPayloadLen - maximum length of message Payload (64 byte packet - 4 bytes header - 2 bytes crc)
const MaxPayloadLen = packetSize - 4 - 2

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

// SizeError is returned, when input data is of invalid size (e.g. message payload for transfer)
type SizeError uint32

func (f SizeError) Error() string {
	return fmt.Sprintf("Size of passed data too large, allowed %v, got %v", MaxPayloadLen, uint32(f))
}

// CommandID - ID of the USB commands
type CommandID uint8

const (
	// CmdTest - test command, do not use, used for tests
	CmdTest CommandID = 0x61
	// CmdIrq - interrupt callback, only from device to host, used for tests
	CmdIrq CommandID = 0x80
)

// the message type represents a message which is built out of the incoming byte stream
type message struct {
	cmd     CommandID
	err     uint8
	payload []byte
}

// the callback type is used by the receive routine to map command IDs to callback functions
type callback struct {
	cmd    CommandID
	cbFunc IncomingMessageCallback
}

// IncomingMessageCallback - function prototype for incoming message callbacks
// When called the function gets passed the error byte of the message and the payload
type IncomingMessageCallback func(err byte, payload []byte)

/////////////////////////////
// Package variables (private)
/////////////////////////////
var crcTable *crc16.Table
var port *serial.Port

var rxChannel chan message           // Used to pass incoming serial messages from the readerThread to the receive goroutine
var ansChannel chan message          // Used to pass incoming serial messages as answer from the the receive goroutine to the transfer function
var regCallbackChannel chan callback // Used to register callbacks in the receive goroutine

/////////////////////////////
// Package API (public)
/////////////////////////////

// Open connects to the specified virtual COM port
// The parameter 'device' holds the name of the device to connect to, i.e. '/dev/ttyACM0'
func Open(device string) error {
	var err error
	// Open port in mode 115200_N81
	c := &serial.Config{Name: device, Baud: 115200}
	port, err = serial.OpenPort(c)

	if err == nil {
		// Start reader goroutine, which sends incoming messages on rxChannel
		rxChannel = make(chan message)
		ansChannel = make(chan message)
		regCallbackChannel = make(chan callback)

		go serialReaderThread()
		go receive()
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
//   cmd - command ID of the transfer
//   payload - payload to transmit, can be nil for zero TX payload (request-only style commands)
// Returnvalues are Answer-ErrorCode, Payload, error
func Transfer(cmd CommandID, payload []byte) (byte, []byte, error) {
	if len(payload) > MaxPayloadLen {
		return 0, nil, SizeError(len(payload))
	}
	var txBuf [packetSize]byte

	txBuf[0] = sync
	txBuf[1] = byte(cmd)
	txBuf[2] = 0

	if payload == nil {
		txBuf[3] = 0
	} else {
		txBuf[3] = byte(len(payload))
		copy(txBuf[4:], payload[:])
	}

	crc := crc16.Checksum(txBuf[:len(txBuf)-2], crcTable)
	var h, l uint8 = uint8(crc & 0xff), uint8(crc >> 8)
	txBuf[62] = byte(h)
	txBuf[63] = byte(l)

	// Send the message
	bytesWritten, err := port.Write(txBuf[:])

	if err != nil {
		return 0, nil, err
	}

	if bytesWritten != len(txBuf) {
		return 0, nil, errors.New("Error writing bytes to serial port")
	}

	// Wait for answer
	answer := <-ansChannel

	// check that answer actually matches request (cmdID)
	if answer.cmd != cmd {
		// Answer command byte must be identical
		return 0, nil, errors.New("Got unexpected Answer (CommandId did not match")
	}

	return answer.err, answer.payload, nil
}

func receive() {
	var callbacks []callback

	for {
		select {
		case tempCallback := <-regCallbackChannel:
			// register callback, add to callbacks list if function is valid
			if tempCallback.cbFunc != nil {
				callbacks = append(callbacks, tempCallback)
			}

		case msg := <-rxChannel:
			isAnswer := true
			// message received, look if a callback is registered
			for _, cb := range callbacks {
				if cb.cmd == msg.cmd {
					cb.cbFunc(msg.err, msg.payload)
					isAnswer = false
					break
				}
			}
			if isAnswer {
				ansChannel <- msg
			}
		}
	}

}

func serialReaderThread() {

	for {
		var rxBuf [packetSize]byte

		if port != nil {

			bytesRead, err := port.Read(rxBuf[:])

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

			// send message to rxChannel
			rxChannel <- message{
				cmd:     CommandID(rxBuf[idxCmd]),
				err:     rxBuf[idxErr],
				payload: rxBuf[idxPayload : idxPayload+payloadLen]}

			break
		}

	}
}

// RegisterCallback registers a function which is called when message with a certain CommandId is incoming
func RegisterCallback(cmd CommandID, cbFunc IncomingMessageCallback) error {

	if cbFunc == nil {
		return errors.New("Callback parameter should not be nil")
	}

	regCallbackChannel <- callback{cmd, cbFunc}

	return nil
}

//////////////////////////////
// Internal functions (private)
//////////////////////////////

func init() {
	// create crc16 table
	crcTable = crc16.MakeTable(crc16.CRC16_CCITT_FALSE)
}
