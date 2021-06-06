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
	clientrecv network.ClientImpl
	clientreq  = make(chan network.ClientRequest)
	clientresp = make(chan network.ClientResponse)
	ip         = flag.String("ip", "127.0.0.1", "Remote server address")
	port       = flag.Int("port", 8080, "Remote server port")
	secret     = flag.String("key", "secret-key", "Client secret key")
)

func callback(out, in []byte, samples int, samplesize int) {
	clientreq <- network.ClientRequest{Samples: samples}
	copy(out, <-clientresp)
}

func main() {
	flag.Parse()
	var err error
	log.Printf("Connecting to [%s:%d]\n", *ip, *port)
	clientrecv = protocol.NewWebsocketClient(*secret)
	//clientrecv = protocol.NewUDPClient(*secret, "SPYA")
	go network.InitClient(clientrecv, network.ClientReceive, *ip, *port, nil, clientreq, clientresp)
	strm, err := audio.NewAudioStream(audio.Playback, 1, 44100)
	if err != nil {
		log.Fatalf("Unable to create audio stream: %v\n", err)
	}
	strm.SetCallback(callback)
	strm.Start()
	fmt.Println("Press Enter to stop receiving audio...")
	fmt.Scanln()
	strm.Close()
	clientrecv.Disconnect()
}
