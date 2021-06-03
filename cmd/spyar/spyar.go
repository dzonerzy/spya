package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/dzonerzy/spyac/pkg/audio"
	"github.com/dzonerzy/spyac/pkg/protocol"
)

var (
	m          *protocol.Message
	clientrecv *protocol.AudioClient
	ip         = flag.String("ip", "127.0.0.1", "Remote server address")
	port       = flag.Int("port", 8080, "Remote server port")
	secret     = flag.String("key", "secret-key", "Client secret key")
)

func callback(out, in []byte, samples int, samplesize int) {
	data, err := clientrecv.Recv(samples)
	if err != nil {
		log.Printf("Trying to reconnect...")
		clientrecv.Reconnect()
	}
	m.Unserialize(data)
	copy(out, m.Data)
}

func main() {
	flag.Parse()
	var err error
	log.Printf("Connecting to [%s:%d]\n", *ip, *port)
	clientrecv, err = protocol.NewAudioClient(*ip, *port, protocol.TypeReceive, *secret)
	if err != nil {
		log.Fatalf("Unable to connect: %v\n", err)
	}
	log.Printf("Connected")
	m = protocol.NewMessage(nil)
	strm, err := audio.NewAudioStream(audio.Playback, 1, 44100)
	if err != nil {
		log.Fatalf("Unable to create audio stream: %v\n", err)
	}
	strm.SetCallback(callback)
	strm.Start()
	fmt.Println("Press Enter to receiving audio...")
	fmt.Scanln()
	strm.Close()
}
