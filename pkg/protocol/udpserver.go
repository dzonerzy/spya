package protocol

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/dzonerzy/spya/pkg/audio"
	"github.com/xtaci/kcp-go"
)

type UDPServer struct {
	block    kcp.BlockCrypt
	listener *kcp.Listener
	running  bool
	lasterr  error
	clients  map[string]*audio.AudioBuffer
}

func (us *UDPServer) createClient(key string) {
	if key != "" {
		if _, ok := us.clients[key]; !ok {
			us.clients[key] = audio.NewAudioBuffer(64 * audio.Kilobyte)
		}
	}
}

func (us *UDPServer) removeClient(key string) {
	if key != "" {
		if _, ok := us.clients[key]; ok {
			us.clients[key] = nil
			delete(us.clients, key)
		}
	}
}

func (us *UDPServer) existsClient(key string) bool {
	if key != "" {
		_, ok := us.clients[key]
		return ok

	}
	return false
}

func (u *UDPServer) handleClient(s *kcp.UDPSession /**smux.Session*/) {
	/*stream, err := s.AcceptStream()
	if err != nil {
		log.Fatal(err)
	}*/
	stream := s
	var m = NewMessage()
	var key = ""
	for {
		packetLength := make([]byte, 4)
		_, err := stream.Read(packetLength)
		if err != nil {
			u.lasterr = err
			if !u.existsClient(key) {
				u.removeClient(key)
			}
			stream.Close()
			break
		}
		length := binary.LittleEndian.Uint32(packetLength)
		var packet = make([]byte, length)
		_, err = stream.Read(packet)
		if err != nil {
			u.lasterr = err
			if !u.existsClient(key) {
				u.removeClient(key)
			}
			stream.Close()
			break
		}
		m.Unserialize(packet)
		if key == "" {
			key = m.Key
			if !u.existsClient(key) {
				u.createClient(key)
			}
		}
		if m.Type == SendData {
			u.clients[key].WriteSamples(m.Data, int(m.Samples))
		} else if m.Type == RecvData {
			if !u.existsClient(m.Key) {
				stream.Close()
				break
			}
			samplesToRead := int(m.Samples)
			data := u.clients[key].ReadSamples(samplesToRead)
			m.Data = data
			m.Type = SendData
			m.Samples = int32(samplesToRead)
			_, err = stream.Write(m.Serialize())
			if err != nil {
				u.lasterr = err
				stream.Close()
				break
			}
		}
	}
}

func (u *UDPServer) Start(ip string, port int) {
	u.listener, _ = kcp.ListenWithOptions(fmt.Sprintf("%s:%d", ip, port), u.block, 10, 3)
	u.running = true
	for u.running {
		s, err := u.listener.AcceptKCP()
		if err != nil {
			u.lasterr = err
			return
		}
		/*sess, err := smux.Server(s, nil)
		if err != nil {
			u.lasterr = err
			return
		}*/
		log.Printf(" Got connection from [%s]\n", s.RemoteAddr().String())
		go u.handleClient(s /*sess*/)
	}
	u.lasterr = nil
}

func (u *UDPServer) Stop() {
	u.lasterr = u.listener.Close()
}

func (u *UDPServer) GetLastError() error {
	return u.lasterr
}

func (u *UDPServer) Loop() {
	var s = make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	<-s
	u.running = false
	u.Stop()
}

func NewUDPServer(key string) *UDPServer {
	block, _ := kcp.NewSalsa20BlockCrypt(sha1.New().Sum([]byte(key)))
	return &UDPServer{
		block:   block,
		clients: make(map[string]*audio.AudioBuffer),
	}
}
