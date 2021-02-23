package esbbridgeclient

import (
	"testing"
)

func TestTemp(t *testing.T) {

	Connect("localhost:9815")

	Disconnect()

}
