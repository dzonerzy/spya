package protocol

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"

	"github.com/dzonerzy/spya/pkg/network"
	"github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

type UDPClient struct {
	block     kcp.BlockCrypt
	conn      *kcp.UDPSession
	stream    *smux.Stream
	clientKey string
}

func (uc *UDPClient) Connect(addr string, port int, clientType network.ClientType) error {
	var err error
	uc.conn, err = kcp.DialWithOptions(fmt.Sprintf("%s:%d", addr, port), uc.block, 10, 3)
	if err != nil {
		return err
	}
	/*session, err := smux.Client(conn, nil)
	if err != nil {
		return err
	}
	// Open a new stream
	uc.stream, err = session.OpenStream()*/
	return err
}

func (uc *UDPClient) SendMessage(data []byte, samples int) error {
	var m = NewMessage()
	m.SetKey(uc.clientKey)
	m.Type = SendData
	m.Data = data
	m.Samples = int32(samples)
	_, err := uc.conn.Write(m.Serialize())
	return err
}

func (uc *UDPClient) ReadMessage(samples int) ([]byte, error) {
	var m = NewMessage()
	m.SetKey(uc.clientKey)
	m.Type = RecvData
	m.Samples = int32(samples)
	_, err := uc.conn.Write(m.Serialize())
	if err != nil {
		return nil, err
	}
	responseLength := make([]byte, 4)
	_, err = uc.conn.Read(responseLength)
	if err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(responseLength)
	var response = make([]byte, length)
	_, err = uc.conn.Read(response)
	if err != nil {
		return nil, err
	}
	m.Unserialize(response)
	return m.Data, nil
}

func (uc *UDPClient) Disconnect() error {
	return uc.stream.Close()
}

func NewUDPClient(clientKey, key string) *UDPClient {
	block, _ := kcp.NewSalsa20BlockCrypt(sha1.New().Sum([]byte(key)))
	return &UDPClient{
		block:     block,
		clientKey: clientKey,
	}
}
