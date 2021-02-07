package usbprotocol

import (
	"fmt"

	"go.bug.st/serial.v1"
)

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
func Transfer(message []byte) {
	fmt.Printf("%s", message)
}

func init() {
	// Todo: do init stuff here
	fmt.Printf("Init")
}
