package main

import (
	"flag"
	"log"

	"github.com/dzonerzy/spya/pkg/network"
	"github.com/dzonerzy/spya/pkg/protocol"
)

var (
	ip   = flag.String("ip", "127.0.0.1", "Remote server address")
	port = flag.Int("port", 8080, "Remote server port")
)

func main() {
	flag.Parse()
	log.Printf("Starting server on [%s:%d]\n", *ip, *port)
	server := protocol.NewWebsocketServer()
	//server := protocol.NewUDPServer("SPYA")
	err := network.StartServer(server, *ip, *port, nil)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	log.Println("Server stopped")
	if err != nil {
		log.Fatal(err)
	}
}
