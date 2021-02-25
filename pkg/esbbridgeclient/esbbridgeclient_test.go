package esbbridgeclient

import (
	"fmt"
	"testing"
)

func TestTemp(t *testing.T) {

	Connect("localhost:9815")
	var testPipelineAddress = []byte{111, 111, 111, 111, 1}

	fmt.Println(Transfer(testPipelineAddress, 0x10, nil))
	Disconnect()

}
