package audio

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
)

type Size int

const (
	Byte     Size = 1
	Kilobyte      = 1024 * Byte
	Megabyte      = 1024 * Kilobyte
)

type AudioBuffer struct {
	maxsize       Size
	buf           *bytes.Buffer
	m             sync.RWMutex
	storedsamples int32
}

func (ab *AudioBuffer) WriteSamples(data []byte, samples int) {
	ab.m.Lock()
	defer ab.m.Unlock()
	var expected = atomic.LoadInt32(&ab.storedsamples) + int32(samples)
	if int(expected) < int(ab.maxsize) {
		ab.buf.Write(data)
		atomic.AddInt32(&ab.storedsamples, int32(samples))
		//ab.storedsamples += samples
	} else {
		d := ab.buf.Bytes()
		d = d[samples:]
		ab.buf.Reset()
		ab.buf.Write(d)
		ab.buf.Write(data)
	}
}

func (ab *AudioBuffer) ReadSamples(samples int) []byte {
	var out = make([]byte, samples)
	for {
		var stored = atomic.LoadInt32(&ab.storedsamples)
		if int(stored) >= samples {
			ab.m.RLock()
			defer ab.m.RUnlock()
			n, _ := ab.buf.Read(out)
			atomic.StoreInt32(&ab.storedsamples, ab.storedsamples-int32(n))
			return out
		}
	}
}

func (ab *AudioBuffer) Size() int {
	fmt.Println(ab.buf.Bytes())
	return int(atomic.LoadInt32(&ab.storedsamples))
}

func NewAudioBuffer(size Size) *AudioBuffer {
	return &AudioBuffer{
		maxsize:       size,
		buf:           new(bytes.Buffer),
		storedsamples: 0,
	}
}
