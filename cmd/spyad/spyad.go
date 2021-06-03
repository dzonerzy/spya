package main

import (
	"log"

	"github.com/dzonerzy/spya/pkg/protocol"
)

func main() {
	server, err := protocol.NewAudioServer("127.0.0.1", 8080)
	if err != nil {
		log.Fatal(err)
	}
	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}
