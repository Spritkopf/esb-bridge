package esbbridgeclient

import (
	"fmt"
	"testing"
)

func TestTemp(t *testing.T) {

	Connect("localhost:9815")

	fmt.Println(Transfer([]byte{6, 4, 3, 2, 1}, 99, []byte{1, 2, 3}))
	Disconnect()

}
