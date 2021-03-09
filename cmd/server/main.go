package main

///////////////////////////////////////////////////////////////////////////////
// ESB bridge server
//
// Console application to start the esb-bridge RPC server

import (
	"log"

	"github.com/alecthomas/kong"
	"github.com/spritkopf/esb-bridge/pkg/server"
)

var opts struct {
	Verbose bool   `short:"v" help:"Additional output"`
	Port    uint   `short:"p" name:"port" default:"9815" help:"TCP port to listen on (default: 9815)"`
	Device  string `short:"d" name:"device" help:"Serial port of the esb-bridge device (e.g. /dev/ttyACM0)"`
}

func main() {
	kong.Parse(&opts)

	cancel, err := server.Start(opts.Device, opts.Port)
	defer cancel()

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	for {
		// Wait here for a SIGINT (CTRL+C)
	}
}
