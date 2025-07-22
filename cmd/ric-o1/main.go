// cmd/ric-o1/main.go
package main

import (
	"log"
	"net"

	"github.com/hctsai1006/near-rt-ric/pkg/o1/netconf"
)

func main() {
	// Create a new NETCONF server
	server, err := netconf.NewServer()
	if err != nil {
		log.Fatalf("Failed to create NETCONF server: %v", err)
	}

	// Listen on port 8300
	listener, err := net.Listen("tcp", ":8300")
	if err != nil {
		log.Fatalf("Failed to listen on port 8300: %v", err)
	}
	defer listener.Close()

	// Start the server
	server.Start(listener)
}
