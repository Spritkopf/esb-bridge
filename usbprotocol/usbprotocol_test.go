package usbprotocol

import (
	"testing"
)

//TestOpenSuccess tests that the virtual COM port can be opened
func TestOpenSuccess(t *testing.T) {
	err := Open(("/dev/ttyACM0"))

	if err != nil {
		t.Fatal(err)
	}

	Close()
}

//TestOpenFail tests the error handling of Open() in case of bad port name
func TestOpenFail(t *testing.T) {
	err := Open(("djkdskfj"))

	if err == nil {
		t.FailNow()
	}
}

// TestTransfer tests the successful operation of the Transfer function
func TestTransfer(t *testing.T) {

	Open("/dev/ttyACM0")

	msg := []byte{1, 2, 3, 4}
	_, err := Transfer(msg)

	if err != nil {
		t.Fatalf(err.Error())
	}

	Close()

}

// TestTransferMessageTooLong tests the error handling of the Transfer function
func TestTransferMessageTooLong(t *testing.T) {

	msg := make([]byte, 64)
	//msg := []byte{'g', 'o', 'a', 'l'}
	_, err := Transfer(msg)

	_, ok := err.(SizeError)
	if !ok {
		t.Fatalf("Expected Transfer to fail because of too large message parameter")
	}

}
