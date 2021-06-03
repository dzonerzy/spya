package audio

import (
	"bytes"
	"sync"
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
	storedsamples int
}

func (ab *AudioBuffer) WriteSamples(data []byte, samples int) {
	ab.m.Lock()
	defer ab.m.Unlock()
	if ab.storedsamples+samples < int(ab.maxsize) {
		ab.buf.Write(data)
		ab.storedsamples += samples
	} else {
		ab.storedsamples = samples
		ab.buf.Reset()
		ab.buf.Write(data)
	}
}

func (ab *AudioBuffer) ReadSamples(samples int) []byte {
	ab.m.RLock()
	defer ab.m.RUnlock()
	var out = make([]byte, samples)
	for {
		if ab.storedsamples >= samples {
			ab.buf.Read(out)
			return out
		}
	}
}

func (as *AudioBuffer) Size() int {
	return as.storedsamples
}

func NewAudioBuffer(size Size) *AudioBuffer {
	return &AudioBuffer{
		maxsize:       size,
		buf:           new(bytes.Buffer),
		storedsamples: 0,
	}
}
