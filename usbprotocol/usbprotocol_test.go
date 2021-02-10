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
	_, ans, err := Transfer(CmdTest, pl)

	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Printf("Answer: %v\n", ans)
	Close()

}

// TestTransferMessageTooLong tests the error handling of the Transfer function
func TestTransferMessageTooLong(t *testing.T) {

	pl := make([]byte, 65)

	_, _, err := Transfer(CmdTest, pl)

	_, ok := err.(SizeError)
	if !ok {
		t.Fatalf("Expected Transfer to fail because of too large message parameter, but it didn't")
	}

}

// TestEcho Sends a USB packet and expects an echo messsage back
func TestEcho(t *testing.T) {

	Open("/dev/ttyACM0")

	msg := []byte{5, 19, 20}
	ans_err, ans_payload, err := Transfer(CmdTest, msg)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if ans_err != 0 {
		t.Fatalf("Unexpected answer Error Code: Expected:0 , Got:%v", ans_err)
	}

	for i := 0; i < len(msg); i++ {
		if msg[i] != ans_payload[i] {
			t.Fatalf("Unexpected answer: Expected:%v , Got:%v", msg, ans_payload)
		}
	}

	Close()

}

// TestMisc is temporary, used to try out stuff during development
func TestMisc(t *testing.T) {

	Open("/dev/ttyACM0")

	ans_err, ans, err := Transfer(CmdVersion, nil)

	if ans_err != 0 {
		t.Fatalf("Unexpected answer Error Code: Expected:0 , Got:%v", ans_err)
	}

	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Printf("Version: %v\n", ans)
	Close()

}
