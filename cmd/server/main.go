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
		header := make([]byte, 2)
		//_, err := io.ReadAtLeast(conn, header, 2)
		_, err := io.ReadFull(conn, header)

		if err != nil {
			fmt.Println("Client disconnected.")
			conn.Close()
			return
		}

		if len(header) < 2 {
			log.Println("Packet too short, need at least 2 bytes (cmd + len)")
			continue
		}
		payloadSize := header[1]
		payload := make([]byte, int(payloadSize))
		if payloadSize > 0 {
			_, err = io.ReadFull(conn, payload)
		}

		log.Printf("Client message: Cmd %v, payload %v", header[0], payload)

		answer := make([]byte, 3, 64)
		answer[0] = header[0] // answer byte 0: same command byte as the request
		answer[1] = 0         // answer byte 1: error code
		// check message content
		switch header[0] {
		// Transfer command
		case 0x00:
			// payload must at least contain the target address and a command
			if payloadSize < 6 {
				log.Println("Transfer: Packet too short, need at least 6 bytes (5 address + 1 cmd)")
				answer[1] = 0x01 // error: payload size

			} else {
				addr := [5]byte{}
				copy(addr[:5], payload[:5])

				log.Printf("ESB Transfer: Addr %v payload: %v", addr, payload[5:])

				esbAnsPayload, err := esbbridge.Transfer(addr, payload[5:])

				if err != nil {
					log.Printf("Tansfer Error: %v", err)
					answer[1] = 0x02 // error: transfer error
				} else {
					answer[2] = uint8(len(esbAnsPayload))
					answer = append(answer, esbAnsPayload...)
				}

			}

		}
		log.Println("Answer: ", answer)
		conn.Write(answer)
	}
}

func exit(errCode int) {
	esbbridge.Close()
	os.Exit(errCode)
}
