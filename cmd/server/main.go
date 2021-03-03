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

// CommandID - ID of TCP commands
type CommandID byte

const (
	// CmdTransfer - transfer an ESB message
	CmdTransfer CommandID = 0x10
	// CmdListen - Start listening for ESB messages
	CmdListen CommandID = 0x20
)

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
	killChannel := make(chan bool, 1)
	for {
		header := make([]byte, 2)
		//_, err := io.ReadAtLeast(conn, header, 2)
		_, err := io.ReadFull(conn, header)

		if err != nil {
			fmt.Println("Client disconnected.")
			killChannel <- true
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
		answer[2] = 0         // answer byte 2: payload length
		// check message content
		switch CommandID(header[0]) {
		// Transfer command
		case CmdTransfer:
			// payload must at least contain the target address and a command
			if payloadSize < 6 {
				log.Println("Transfer: Packet too short, need at least 6 bytes (5 address + 1 cmd)")
				answer[1] = 0x01 // error: payload size
				break
			}

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

		case CmdListen:
			// payload must contain the target address and a command
			if payloadSize != 6 {
				log.Println("Transfer: Invalid packet, must be 6 bytes long (5 address + 1 cmd)")
				answer[1] = 0x01 // error: payload size
				break
			}
			listenAddr := [5]byte{}
			copy(listenAddr[:5], payload[:5])
			listenCmd := payload[5]

			lc := make(chan esbbridge.EsbMessage, 1)
			esbbridge.AddListener(listenAddr, listenCmd, lc)

			go func(conn net.Conn, lc chan esbbridge.EsbMessage, killChannel chan bool) {
				for {
					select {
					case msg := <-lc:
						log.Printf("Message received: %v\n", msg)
						txMsg := make([]byte, 3, 64)
						txMsg[0] = 0x21 // todo, replace magic numbers with constants
						txMsg[1] = 0x00
						txMsg[2] = uint8(6 + len(msg.Payload))
						txMsg = append(txMsg, msg.Address[:]...)
						txMsg = append(txMsg, msg.Cmd)
						txMsg = append(txMsg, msg.Payload...)
						log.Printf("Send: %v\n", txMsg)
						conn.Write(txMsg)
					case <-killChannel:
						esbbridge.RemoveListener(lc)
						log.Printf("Removing listener\n")
					}
				}
			}(conn, lc, killChannel)

			log.Printf("Adding listener for address %v, Cmd %v\n", listenAddr, listenCmd)

		case 0x88:
			answer[2] = header[1]
			answer = append(answer, payload...)
			conn.Write(answer)
			fmt.Printf("dummy command, ECHO %v\n", answer)
		}
		log.Printf("Send Answer to client [%v]: %v\n", conn.RemoteAddr(), answer)
		conn.Write(answer)

	}
}

func exit(errCode int) {
	esbbridge.Close()
	os.Exit(errCode)
}
