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
		t.Fatal(err)
	}
	Close()
}

// TestTransferMessageTooLong tests the error handling of the Transfer function when tx payload is too long
func TestTransferMessageTooLong(t *testing.T) {

	var expectedErrCode = ErrSize.ErrCode
	pl := make([]byte, 65)

	msg := Message{Cmd: CmdTest, Payload: pl}
	_, err := Transfer(msg)

	e, ok := err.(UsbError)
	if (!ok) || (e.ErrCode != expectedErrCode) {
		t.Fatalf("Expected ErrSize (%v), got: %v", expectedErrCode, e)
	}

}

// TestTransfer tests the successful operation of the Transfer function
func TestTransfer(t *testing.T) {

	Open("/dev/ttyACM0")

	pl := []byte{1, 2, 3, 4}
	msg := Message{Cmd: CmdTest, Payload: pl}
	_, err := Transfer(msg)

	if err != nil {
		t.Fatalf(err.Error())
	}

	//fmt.Printf("Answer: %v\n", ans)

	Close()

}

// TestTransfer tests the successful tarnsfer of multiple messages in quick succession
func TestTransferMulti(t *testing.T) {

	Open("/dev/ttyACM0")
	pl := []byte{1, 2, 3, 4}
	msg := Message{Cmd: CmdTest, Payload: pl}
	for i := 0; i < 5; i++ {
		_, err := Transfer(msg)

		if err != nil {
			t.Fatalf(err.Error())
		}

		//fmt.Printf("Answer %v: %v Err %v\n", i, ans.Payload, ans.Err)
	}
	Close()

}

// TestTransferInvalidCommand tests the error handling on invalid command ID
func TestTransferInvalidCommand(t *testing.T) {

	Open("/dev/ttyACM0")

	msg := Message{Cmd: 0xFE, Payload: nil}
	answer, err := Transfer(msg)

	if answer.Err != 0x10 {
		t.Fatalf("Answer message should have the E_NO_CMD Error code when requesting a unknown command")
	}

	if err != nil {
		t.Fatalf(err.Error())
	}

	Close()
}

// TestTransferTimeout tests the error handling on timeout while waiting for a response from the device
func TestTransferTimeout(t *testing.T) {

	///////////
	// Manual Test: Uncomment below and run the test manually
	//////////

	// var expectedErrCode = ErrTimeout.ErrCode
	// Open("/dev/ttyACM0")

	// _, _, err := Transfer(CmdTest, nil)

	// e, ok := err.(UsbError)
	// if (!ok) || (e.ErrCode != expectedErrCode) {
	// 	t.Fatalf("Expected ErrTimeout (%v), got: %v", expectedErrCode, e)
	// }

	// Close()
}

func TestRegisterCallbackFailed(t *testing.T) {
	var expectedErrCode = ErrParam.ErrCode

	err := RegisterCallback(CmdIrq, nil)

	e, ok := err.(UsbError)
	if (!ok) || (e.ErrCode != expectedErrCode) {
		t.Fatalf("Expected ErrParam (%v), got: %v", expectedErrCode, e)
	}
}

// TestCallback tests that a registered callback is called
// Note: This is a manual test, it requires the user to press a button on the board
func TestCallback(t *testing.T) {

	///////////
	// Manual Test: Uncomment below and run the test manually
	//////////

	// messageReceived := false

	// Open("/dev/ttyACM0")
	// RegisterCallback(CmdIrq, func(err byte, payload []byte) {
	// 	fmt.Printf("Payload: %v", payload)
	// 	messageReceived = true
	// })
	// fmt.Printf("Please press the button during the next 60 seconds\n")
	// for i := 10; i > 0; i-- {
	// 	if messageReceived {
	// 		break
	// 	}
	// 	fmt.Printf("%v\n", i)
	// 	time.Sleep(1 * time.Second)
	// }

	// Close()

	// if !messageReceived {
	// 	t.Fatalf("Timeout, no message was received")
	// }

}

// TestEcho Sends a USB packet and expects an echo messsage back
func TestEcho(t *testing.T) {

	Open("/dev/ttyACM0")

	pl := []byte{5, 19, 20}
	msg := Message{Cmd: CmdTest, Payload: pl}
	answer, err := Transfer(msg)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if answer.Err != 0 {
		t.Fatalf("Unexpected answer Error Code: Expected:0 , Got:%v", answer.Err)
	}

	for i := 0; i < len(pl); i++ {
		if pl[i] != answer.Payload[i] {
			t.Fatalf("Unexpected answer: Expected:%v , Got:%v", msg, answer.Payload)
		}
	}

	Close()

}
