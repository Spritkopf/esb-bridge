package main

///////////////////////////////////////////////////////////////////////////////
// ESB bridge server
//
// Provides access to a connected esb-bridge device over a TCP socket
//
// Main Features:
// - transparent communication with the esb-bridge device
// - setup listeners for arbitrary packets

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/spritkopf/esb-bridge/pkg/esbbridge"
)

const serverVersion string = "0.1.0"

var opts struct {
	Verbose bool   `short:"v" help:"Additional output"`
	Version bool   `name:"version" help:"Print version and exit"`
	Port    uint   `short:"p" name:"port" default:"9815" help:"TCP port to listen on (default: 9815)"`
	Device  string `short:"d" name:"device" help:"Serial port of the esb-bridge device (e.g. /dev/ttyACM0)"`
}

var socket net.Listener

func main() {
	kong.Parse(&opts)

	// check flag: version
	if opts.Version {
		fmt.Printf("esb-bridge-server version %v\n", serverVersion)
		os.Exit(0)
	}

	// setup a handler for SIGINT (ctrl+c)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n- Received SIGINT (Ctrl+C). Exiting...")
		exit(0)
	}()

	err := esbbridge.Open(opts.Device)
	defer exit(0)

	if err != nil {
		panic(err)
	}

	fwVersion, err := esbbridge.GetFwVersion()
	if err != nil {
		log.Printf("Error: Could not get version information: %v", err)
		exit(1)
	}

	log.Printf("Connected to esb-bridge-fw, version: %v\n", fwVersion)

	log.Printf("Start listening on port %v\n", opts.Port)

	socket, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", opts.Port))
	if err != nil {
		fmt.Println("Error opening TCP socket:", err.Error())
		exit(1)
	}
	defer socket.Close()

	for {
		c, err := socket.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		log.Println("Client connected.")

		log.Println("Client " + c.RemoteAddr().String() + " connected.")

		go handleConnection(c)
	}
}

func handleConnection(conn net.Conn) {
	for {
		buffer := make([]byte, 2, 64)
		_, err := io.ReadAtLeast(conn, buffer, 2)

		if err != nil {
			fmt.Println("Client disconnected.")
			conn.Close()
			return
		}

		log.Println("Client message:", buffer)

		if len(buffer) < 2 {
			log.Println("Packet too short, need at least 2 bytes (cmd + len)")
			continue
		}
		payloadSize := buffer[1]
		payload := buffer[2:]

		answer := make([]byte, 2, 64)
		answer[0] = buffer[0] // answer byte 0: same command byte as the request
		answer[1] = 0         // answer byte 1: error code
		// check message content
		switch buffer[0] {
		// Transfer command
		case 0x00:
			// payload must at least contain the target address and a command
			if payloadSize < 6 {
				log.Println("Transfer: Packet too short, need at least 6 bytes (5 address + 1 cmd)")
				answer[1] = 0x01 // error: payload size
			} else {
				addr := [5]byte{}
				copy(addr[:5], payload[:5])

				//esbAnsPayload, err := esbbridge.Transfer(addr, payload[5:])
				esbAnsPayload := buffer
				if err != nil {
					answer[1] = 0x02 // error: transfer error
				} else {
					answer[2] = uint8(len(esbAnsPayload))
					copy(answer[2:], esbAnsPayload[:])
				}

			}

		}
		conn.Write(answer)
	}
}

func exit(errCode int) {
	esbbridge.Close()
	os.Exit(errCode)
}
