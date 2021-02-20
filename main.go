package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
)

var opts struct {
	Verbose bool `help:"Additional output"`
	Port    uint `short:"p" name:"port" default:"9815" help:"TCP port to listen on (default: 9815)"`
}

func main() {
	kong.Parse(&opts)

	// setup a handler for SIGINT (ctrl+c)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n- Received SIGINT (Ctrl+C). Exiting...")
		os.Exit(0)
	}()

	fmt.Printf("Start listening on port %v\n", opts.Port)

	l, err := net.Listen("tcp", ":9815")
	if err != nil {
		fmt.Println("Error opening TCP socket:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		fmt.Println("Client connected.")

		fmt.Println("Client " + c.RemoteAddr().String() + " connected.")
	}
}
