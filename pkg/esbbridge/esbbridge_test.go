package esbbridge

import (
	"fmt"
	"testing"
	"time"
)

var testPipelineAddress = [5]byte{111, 111, 111, 111, 1}
var testDevice string = "/dev/ttyACM0"

//TestOpenSuccess tests that the virtual COM port can be opened
func TestOpenSuccess(t *testing.T) {
	err := Open("/dev/ttyACM0")

	if err != nil {
		t.Fatalf(err.Error())
	}

	Close()
}

// TestGetFwVersionNotOpen tests error handling when not connected
func TestGetFwVersionNotOpen(t *testing.T) {

	_, err := GetFwVersion()

	if err == nil {
		t.Fatalf("GetFwVersion should return an error when not connected (i.e. Open() was not called beforehand)")
	}

	Close()

}

// TestGetFwVersion tests correct read of firmware version
func TestGetFwVersion(t *testing.T) {

	err := Open(testDevice)
	defer Close()
	if err != nil {
		t.Fatalf(err.Error())
	}

	version, err := GetFwVersion()

	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Printf("Version: %v\n", version)

}

// TestTransferNotOpen tests error handling when not connected
func TestTransferNotOpen(t *testing.T) {
	_, err := Transfer(EsbMessage{})

	if err == nil {
		t.Fatalf("Transfer should return an error when not connected (i.e. Open() was not called beforehand)")
	}

	Close()
}

// TestTransferPayloadSize tests error handling for too large and too short payloads
func TestTransferPayloadSize(t *testing.T) {
	var veryLongPayload [64]byte

	Open(testDevice)

	_, err := Transfer(EsbMessage{Address: testPipelineAddress[:], Payload: veryLongPayload[:]})

	if err == nil {
		t.Fatalf("Transfer should return an error when Payload is longer than 32 bytes")
	}

	Close()
}

// TestTransfer tests the transfer of ESB packages by requesting the firware version of a supported device
// Note: the ESB command ID ESB_CMD_VERSION (0x10) should be common to all the custom esb compatible devices
func TestTransfer(t *testing.T) {

	errOpen := Open(testDevice)

	if errOpen != nil {
		t.Fatalf("Open() failed with error %v", errOpen)
	}

	answer, err := Transfer(EsbMessage{Address: testPipelineAddress[:], Cmd: 0x10})

	if err != nil {
		t.Fatalf("Transfer() failed with error %v", err)
	}
	fmt.Printf("Answer: %s\n", answer)

	if (answer.Error) != 0 {
		t.Fatalf("Answer Message has error code %v", answer.Error)
	}
	Close()
}

// TestListenerInvalidParam tests that Addlistener will return an error if an invalid channel parameter (nil) is passed
func TestListenerInvalidParam(t *testing.T) {

	err := AddListener([5]byte{}, 0, nil)

	if err == nil {
		t.Fatalf("AddListener should return an error if nil is passed as channel")
	}
}

// TestListener checks that incoming messages can be received
// This is a manual test as it requires a device to send a message
func TestListener(t *testing.T) {
	messageReceived := false

	Open("/dev/ttyACM0")
	defer Close()

	lc := make(chan EsbMessage, 1)

	AddListener([5]byte{12, 13, 14, 15, 16}, 0xFF, lc)

timeoutLoop:
	for i := 10; i > 0; i-- {
		select {
		case msg := <-lc:
			fmt.Printf("Message received: %v", msg)
			messageReceived = true
			break timeoutLoop
		case <-time.After(1 * time.Second):
			fmt.Printf("%v\n", i)
		}
	}

	if !messageReceived {
		t.Fatalf("Timeout, no message was received")
	}
}

// TestRemoveListener tests removal of listeners
func TestRemoveListener(t *testing.T) {
	lc := make(chan EsbMessage, 1)
	lc2 := make(chan EsbMessage, 1)
	AddListener([5]byte{12, 13, 14, 15, 16}, 0x01, lc)
	AddListener([5]byte{12, 13, 14, 15, 16}, 0x02, lc2)
	AddListener([5]byte{12, 13, 14, 15, 16}, 0x02, lc)
	AddListener([5]byte{12, 13, 14, 15, 16}, 0xFF, lc2)

	n := RemoveListener(lc2)

	if n != 2 {
		t.Fatalf("Only two listeners should be removed, but it were %v", n)
	}
}

func TestTemp(t *testing.T) {

}
