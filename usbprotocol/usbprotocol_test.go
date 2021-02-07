package usbprotocol

import (
	"testing"
)

// Test
func TestHelloName(t *testing.T) {
	a := []byte{'g', 'o', 'a', 'l'}

	Transfer(a)

}

func TestOpenSuccess(t *testing.T) {
	err := Open(("/dev/ttyACM0a"))

	if err != nil {
		t.Fatal(err)
	}

	Close()
}

func TestOpenFail(t *testing.T) {
	err := Open(("djkdskfj"))

	if err == nil {
		t.FailNow()
	}
}
