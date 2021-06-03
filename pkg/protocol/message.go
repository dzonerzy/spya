package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"

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

func (m *Message) Unserialize(data []byte) {
	packer := binpacker.NewUnpacker(binary.LittleEndian, bytes.NewReader(data))
	packer.FetchInt32(&m.Samples)
	fmt.Println(m.Samples)
}

func NewMessage(data []byte) *Message {
	return &Message{
		Samples: int32(len(data)),
		Data:    data,
	}
}
