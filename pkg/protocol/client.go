package protocol

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type ConnectionType int

const (
	TypeSend ConnectionType = iota + 1
	TypeReceive
)

type AudioClient struct {
	dialer *websocket.Dialer
	conn   *websocket.Conn
	ctype  ConnectionType
	addr   string
	header http.Header
}

func (ac *AudioClient) Send(data []byte) (err error) {
	err = nil
	if ac.ctype == TypeSend {
		err = ac.conn.WriteMessage(websocket.BinaryMessage, data)
	}
	return
}

func (ac *AudioClient) Recv(samples int) (data []byte, err error) {
	err = nil
	data = nil
	if ac.ctype == TypeReceive {
		var buf = make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(samples))
		err = ac.conn.WriteMessage(websocket.BinaryMessage, buf)
		if err != nil {
			return
		}
		_, data, err = ac.conn.ReadMessage()
	}
	return
}

func (ac *AudioClient) Reconnect() {
	for {
		var err error
		ac.conn, _, err = ac.dialer.Dial(ac.addr, ac.header)
		if err == nil {
			break
		}
	}
}

func (ac *AudioClient) Close() (err error) {
	err = ac.conn.Close()
	return
}

func NewAudioClient(addr string, port int, ctype ConnectionType, secret string) (ac *AudioClient, err error) {
	ac = &AudioClient{
		ctype: ctype,
	}
	ac.dialer = &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  5 * time.Second,
		EnableCompression: true,
	}
	var r *http.Response
	var header = make(http.Header)
	header.Add("Client-Secret", secret)
	ac.header = header
	switch ac.ctype {
	case TypeSend:
		ac.addr = fmt.Sprintf("ws://%s:%d/send", addr, port)
		ac.conn, r, err = ac.dialer.Dial(ac.addr, ac.header)
	case TypeReceive:
		ac.addr = fmt.Sprintf("ws://%s:%d/recv", addr, port)
		ac.conn, r, err = ac.dialer.Dial(ac.addr, ac.header)
	}
	if r.StatusCode == http.StatusNotFound {
		err = fmt.Errorf("client key invalid or not found")
	}
	if err != nil {
		ac = nil
	}
	return
}
