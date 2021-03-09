package client

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	pb "github.com/spritkopf/esb-bridge/pkg/server/service"
)

var (
	serverAddr = flag.String("server_addr", "localhost:9815", "The server address in the format of host:port")
)

func setup() {
	err := Connect(*serverAddr)
	if err != nil {
		log.Fatalf("Setup: Connection Error: %v", err)
	}
}
func teardown() {
	err := Disconnect()
	if err != nil {
		fmt.Printf("Error while disconnection: %v)", err)

	}
}
func TestTransfer(t *testing.T) {

	answerMsg, err := Transfer(client, &pb.EsbMessage{Addr: []byte{111, 111, 111, 111, 1}, Cmd: []byte{0x10}})

	if err != nil {
		t.Fatalf("Transfer returned error: %v", err)
	}

	if answerMsg.Error != 0 {
		t.Fatalf("ESB Answer has Error Code : %v", answerMsg.Error)
	}

	fmt.Printf("Got answer: %v\n", answerMsg)
}

func TestListen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rxChan, _ := Listen(ctx, client, &pb.Listener{Addr: []byte{12, 13, 14, 15, 16}, Cmd: []byte{0xFF}})

	for i := 0; i < 4; i++ {
		msg := <-rxChan
		log.Printf("Incoming Message: %v", msg)
	}
	cancel()
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}
