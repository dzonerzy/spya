package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/zhuangsirui/binpacker"
)

type MessageType uint8

const (
	SendData MessageType = iota + 1
	RecvData
)

type Message struct {
	Type      MessageType
	KeyLength int32
	Key       string
	Samples   int32
	Data      []byte
}

func (m *Message) IsSend() bool {
	return m.Type == SendData
}

func (m *Message) IsRecv() bool {
	return m.Type == RecvData
}

func (m *Message) SetKey(key string) {
	m.Key = key
	m.KeyLength = int32(len(m.Key))
}

func (m *Message) Serialize() []byte {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.LittleEndian, buffer)
	switch m.Type {
	case RecvData:
		packer.PushInt32(9 + int32(len(m.Key)))
	case SendData:
		packer.PushInt32(9 + int32(len(m.Key)) + int32(len(m.Data)))
	}
	packer.PushByte(byte(m.Type)).PushInt32(int32(m.KeyLength)).PushBytes([]byte(m.Key)).PushInt32(int32(m.Samples))
	if m.Type == SendData {
		packer.PushBytes(m.Data)
	}
	return buffer.Bytes()
}

func (m *Message) FromData(data []byte, samples int) {
	m.Type = SendData
	m.Samples = int32(samples)
	m.Data = data[:m.Samples]
}

func (m *Message) Unserialize(data []byte) {
	packer := binpacker.NewUnpacker(binary.LittleEndian, bytes.NewReader(data))
	var t byte
	packer.FetchByte(&t)
	m.Type = MessageType(t)
	var kl int32
	packer.FetchInt32(&kl)
	m.KeyLength = int32(kl)
	var k = make([]byte, m.KeyLength)
	packer.FetchBytes(uint64(m.KeyLength), &k)
	m.Key = string(k)
	packer.FetchInt32(&m.Samples)
	if m.Type == SendData {
		packer.FetchBytes(uint64(m.Samples), &m.Data)
	}
}

func NewMessage() *Message {
	var m = new(Message)
	return m
}
