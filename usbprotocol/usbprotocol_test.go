package usbprotocol

import (
	"fmt"
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

	pl := []byte{1, 2, 3, 4}
	ans, err := Transfer(CmdTest, pl)

	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Printf("Answer: %v\n", ans)
	Close()

}

// TestTransferMessageTooLong tests the error handling of the Transfer function
func TestTransferMessageTooLong(t *testing.T) {

	pl := make([]byte, 65)

	_, err := Transfer(CmdTest, pl)

	_, ok := err.(SizeError)
	if !ok {
		t.Fatalf("Expected Transfer to fail because of too large message parameter, but it didn't")
	}

}

// TestMisc is temporary, used to try out stuff during development
func TestTMisc(t *testing.T) {

	Open("/dev/ttyACM0")

	ans, err := Transfer(CmdVersion, nil)

	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Printf("Version: %v\n", ans[4:7])
	Close()

}
