package esbbridge

import (
	"fmt"
	"testing"
)

var testPipelineAddress = [5]byte{111, 111, 111, 111, 1}

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

	err := Open("/dev/ttyACM0")

	if err != nil {
		t.Fatalf(err.Error())
	}

	version, err := GetFwVersion()

	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Printf("Version: %v\n", version)
	Close()

}

// TestTransferNotOpen tests error handling when not connected
func TestTransferNotOpen(t *testing.T) {
	_, err := Transfer(testPipelineAddress, nil)

	if err == nil {
		t.Fatalf("Transfer should return an error when not connected (i.e. Open() was not called beforehand)")
	}

	Close()
}

// TestTransferPayloadTooLong tests error handling on too large payload
func TestTransferPayloadTooLong(t *testing.T) {
	var veryLongPayload [64]byte

	_, err := Transfer(testPipelineAddress, veryLongPayload[:])

	if err == nil {
		t.Fatalf("Transfer should return an error when Payload is longer than 32 bytes")
	}

	Close()
}
