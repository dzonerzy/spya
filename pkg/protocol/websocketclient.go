package protocol

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dzonerzy/spya/pkg/network"
	"github.com/gorilla/websocket"
)

type ConnectionType int

type WebsocketClient struct {
	dialer *websocket.Dialer
	conn   *websocket.Conn
	ctype  network.ClientType
	secret string
}

func (wsc *WebsocketClient) SendMessage(data []byte, samples int) (err error) {
	err = nil
	var sendmsg = NewMessage()
	if wsc.ctype == network.ClientSend {
		sendmsg.SetKey(wsc.secret)
		sendmsg.Type = SendData
		sendmsg.Samples = int32(samples)
		sendmsg.Data = data
		err = wsc.conn.WriteMessage(websocket.BinaryMessage, sendmsg.Serialize())
	}
	return
}

func (wsc *WebsocketClient) ReadMessage(samples int) (data []byte, err error) {
	err = nil
	data = nil
	var recvmsg = NewMessage()
	if wsc.ctype == network.ClientReceive {
		recvmsg.SetKey(wsc.secret)
		recvmsg.Type = RecvData
		recvmsg.Samples = int32(samples)
		err = wsc.conn.WriteMessage(websocket.BinaryMessage, recvmsg.Serialize())
		if err != nil {
			return
		}
		_, data, err = wsc.conn.ReadMessage()
		if err != nil {
			return nil, err
		}
		recvmsg.Unserialize(data[4:])
		if recvmsg.Type == SendData {
			return recvmsg.Data, nil
		}
	}
	return
}

func (wsc *WebsocketClient) Connect(addr string, port int, ctype network.ClientType) (err error) {
	wsc.ctype = ctype
	switch wsc.ctype {
	case network.ClientSend:
		wsc.conn, _, err = wsc.dialer.Dial(fmt.Sprintf("ws://%s:%d/send", addr, port), nil)
	case network.ClientReceive:
		wsc.conn, _, err = wsc.dialer.Dial(fmt.Sprintf("ws://%s:%d/recv", addr, port), nil)
	}
	return
}

func (wsc *WebsocketClient) Disconnect() (err error) {
	err = wsc.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	err = wsc.conn.Close()
	return
}

func NewWebsocketClient(secret string) (wsc *WebsocketClient) {
	wsc = &WebsocketClient{
		secret: secret,
	}
	wsc.dialer = &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  5 * time.Second,
		EnableCompression: true,
	}
	return
}
