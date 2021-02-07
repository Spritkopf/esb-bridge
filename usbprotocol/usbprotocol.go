package usbprotocol

import (
	"fmt"

	"github.com/sigurn/crc16"
	"go.bug.st/serial.v1"
)

// SizeError is returned, when input data is of invalid size (e.g. message payload for transfer)
type SizeError uint32

func (f SizeError) Error() string {
	return fmt.Sprintf("Size of passed data too large, allowed %v, got %v", MaxPayloadLen, uint32(f))
}

// MaxPayloadLen - maximum length of message Payload (64 byte packet - 4 bytes header - 2 bytes crc)
const MaxPayloadLen = 64 - 4 - 2

// sync byte, marks the beginning of a new packet
const sync = 0x69

const (
	// CmdVersion - Get firmware version
	CmdVersion = 0x10
	// CmdTransfer - Send a message, wait for reply
	CmdTransfer = 0x30
	// CmdSend - Send a message without reply
	CmdSend = 0x31
	// CmdTest - test command, do not use
	CmdTest = 0x61
	// CmdIrq - interrupt callback, only from device to host
	CmdIrq = 0x80
	// CmdRx - callback from incoming ESB message
	CmdRx = 0x81
)

var crcTable *crc16.Table
var port serial.Port

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
func Transfer(message []byte) ([]byte, error) {
	if len(message) > MaxPayloadLen {
		return nil, SizeError(len(message))
	}

	fmt.Printf("%s", message)

	var txBuf [64]byte
	txBuf[0] = sync
	txBuf[1] = CmdTest
	txBuf[2] = 0
	txBuf[3] = byte(len(message))
	copy(txBuf[4:], message[:])
	crc := crc16.Checksum(txBuf[:len(txBuf)-2], crcTable)
	var h, l uint8 = uint8(crc & 0xff), uint8(crc >> 8)
	txBuf[62] = byte(h)
	txBuf[63] = byte(l)

	bytesWritten, err := port.Write(txBuf[:])

	fmt.Println(bytesWritten)
	fmt.Println(err)
	// if err != nil || bytesWritten != len(txBuf) {
	// 	return nil, err
	// }

	return nil, nil
}

func init() {
	// create crc16 table
	crcTable = crc16.MakeTable(crc16.CRC16_CCITT_FALSE)
	//fmt.Printf("Init")
}
