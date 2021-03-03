package esbbridgeclient

import (
	"fmt"
	"testing"
	"time"

	"github.com/spritkopf/esb-bridge/pkg/esbbridge"
)

func TestTransfer(t *testing.T) {

}

func TestListener(t *testing.T) {

	Connect("localhost:9815")
	//var testPipelineAddress = []byte{111, 111, 111, 111, 1}
	var testPipelineAddress = [5]byte{12, 13, 14, 15, 16}

	lc := make(chan esbbridge.EsbMessage, 1)

	AddListener(testPipelineAddress, 0xFF, lc)

	messageReceived := false
timeoutLoop:
	for i := 10; i > 0; i-- {
		select {
		case msg := <-lc:
			fmt.Printf("Test: Message received: %v\n", msg)
			messageReceived = true
			break timeoutLoop
		case <-time.After(1 * time.Second):
			fmt.Printf("%v\n", i)
		}
	}

	if !messageReceived {
		t.Fatalf("Timeout, no message was received")
	}
	Disconnect()

}
