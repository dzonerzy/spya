package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/zhuangsirui/binpacker"
)

type Message struct {
	Samples int32
	Data    []byte
}

func (m *Message) Serialize() []byte {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.LittleEndian, buffer)
	packer.PushInt32(int32(m.Samples)).PushBytes(m.Data)
	return buffer.Bytes()
}

func (m *Message) FromData(data []byte, samples int) {
	m.Data = data
	m.Samples = int32(samples)
}

func (m *Message) Unserialize(data []byte) {
	packer := binpacker.NewUnpacker(binary.LittleEndian, bytes.NewReader(data))
	packer.FetchInt32(&m.Samples)
	packer.FetchBytes(uint64(m.Samples), &m.Data)
}

func NewMessage(data []byte) *Message {
	var m = new(Message)
	m.FromData(data, len(data))
	return m
}
