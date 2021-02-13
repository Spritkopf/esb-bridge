package usbprotocol

import (
	"fmt"
	"testing"
	"time"
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

func TestRegisterCallbackFailed(t *testing.T) {
	err := RegisterCallback(CmdIrq, nil)

	if err == nil {
		t.Fatalf("Passing nil as callback should throw an error")
	}
}

// TestCallback tests that a registered callback is called
// Note: This is a manual test, it requires the user to press a button on the board
func TestCallback(t *testing.T) {

	messageReceived := false

	Open("/dev/ttyACM0")
	RegisterCallback(CmdIrq, func(err byte, payload []byte) {
		fmt.Printf("Payload: %v", payload)
		messageReceived = true
	})
	fmt.Printf("Please press the button during the next 60 seconds\n")
	for i := 10; i > 0; i-- {
		if messageReceived {
			break
		}
		fmt.Printf("%v\n", i)
		time.Sleep(1 * time.Second)
	}

	Close()

	if !messageReceived {
		t.Fatalf("Timeout, no message was received")
	}

}

// TestEcho Sends a USB packet and expects an echo messsage back
func TestEcho(t *testing.T) {

	Open("/dev/ttyACM0")

	msg := []byte{5, 19, 20}
	ansErr, ansPayload, err := Transfer(CmdTest, msg)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if ansErr != 0 {
		t.Fatalf("Unexpected answer Error Code: Expected:0 , Got:%v", ansErr)
	}

	for i := 0; i < len(msg); i++ {
		if msg[i] != ansPayload[i] {
			t.Fatalf("Unexpected answer: Expected:%v , Got:%v", msg, ansPayload)
		}
	}

	Close()

}

// TestMisc is temporary, used to try out stuff during development
func TestMisc(t *testing.T) {

	Open("/dev/ttyACM0")

	ansErr, answer, err := Transfer(CmdVersion, nil)

	if ansErr != 0 {
		t.Fatalf("Unexpected answer Error Code: Expected:0 , Got:%v", ansErr)
	}

	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Printf("Version: %v\n", answer)
	Close()

}
