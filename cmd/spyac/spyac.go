package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/dzonerzy/spya/pkg/audio"
	"github.com/dzonerzy/spya/pkg/network"
	"github.com/dzonerzy/spya/pkg/protocol"
)

var (
	clientdata = make(chan network.ClientData)
	clientsend network.ClientImpl
	ip         = flag.String("ip", "127.0.0.1", "Remote server address")
	port       = flag.Int("port", 8080, "Remote server port")
	secret     = flag.String("key", "secret-key", "Client secret key")
)

func callback(out, in []byte, samples int, samplesize int) {
	clientdata <- network.ClientData{Data: in, Samples: samples}
}

func main() {
	flag.Parse()
	var err error
	log.Printf("Connecting to [%s:%d]\n", *ip, *port)
	clientsend = protocol.NewWebsocketClient(*secret)
	//clientsend = protocol.NewUDPClient(*secret, "SPYA")
	go network.InitClient(clientsend, network.ClientSend, *ip, *port, clientdata, nil, nil)
	strm, err := audio.NewAudioStream(audio.Capture, 1, 44100)
	if err != nil {
		log.Fatalf("Unable to create audio stream: %v\n", err)
	}
	strm.SetCallback(callback)
	strm.Start()
	fmt.Println("Press Enter to stop recording...")
	fmt.Scanln()
	strm.Close()
	clientsend.Disconnect()
}
