package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/dzonerzy/spya/pkg/audio"
	"github.com/dzonerzy/spya/pkg/protocol"
)

var (
	m          *protocol.Message
	clientsend *protocol.AudioClient
	ip         = flag.String("ip", "127.0.0.1", "Remote server address")
	port       = flag.Int("port", 8080, "Remote server port")
	secret     = flag.String("key", "secret-key", "Client secret key")
)

func callback(out, in []byte, samples int, samplesize int) {
	m.FromData(in, samples)
	err := clientsend.Send(m.Serialize())
	if err != nil {
		clientsend.Reconnect()
	}
}

func main() {
	flag.Parse()
	var err error
	log.Printf("Connecting to [%s:%d]\n", *ip, *port)
	clientsend, err = protocol.NewAudioClient(*ip, *port, protocol.TypeSend, *secret)
	if err != nil {
		log.Fatalf("Unable to connect: %v\n", err)
	}
	log.Printf("Connected")
	m = protocol.NewMessage(nil)
	strm, err := audio.NewAudioStream(audio.Capture, 1, 44100)
	if err != nil {
		log.Fatalf("Unable to create audio stream: %v\n", err)
	}
	strm.SetCallback(callback)
	strm.Start()
	fmt.Println("Press Enter to stop recording...")
	fmt.Scanln()
	strm.Close()
}
