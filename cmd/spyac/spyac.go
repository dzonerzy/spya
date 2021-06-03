package main

import (
	"fmt"
	"log"

	"github.com/dzonerzy/spyac/pkg/audio"
	"github.com/dzonerzy/spyac/pkg/protocol"
)

var buffer *audio.AudioBuffer
var m *protocol.Message

func callback(out, in []byte, samples int, samplesize int) {
	//fmt.Println(len(in), samples, samplesize)
	buffer.WriteSamples(in, samples)
	sampledata := buffer.ReadSamples(samples)
	m.Samples = int32(samples)
	m.Data = sampledata
	m.Unserialize(m.Serialize())
	copy(out, m.Data)
}

func main() {
	m = protocol.NewMessage(nil)
	buffer = audio.NewAudioBuffer(64 * audio.Kilobyte)
	strm, err := audio.NewAudioStream(audio.Duplex, 1, 44100)
	if err != nil {
		log.Fatal(err)
	}
	strm.SetCallback(callback)
	strm.Start()
	fmt.Scanln()
	strm.Close()
}
