package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

type UDPServer struct {
	addr     string
	port     int
	listener *kcp.Listener
	sess     *smux.Session
}

func (u *UDPServer) handleClient(s *smux.Session) {
	stream, _ := s.AcceptStream()
	buf := make([]byte, 4096)
	for {
		n, _ := stream.Read(buf)
		stream.Write(buf[:n])
	}
}

func (u *UDPServer) Start() error {
	for {
		s, err := u.listener.AcceptKCP()
		if err != nil {
			return err
		}
		sess, err := smux.Server(s, nil)
		go u.handleClient(sess)
	}
}

func NewUDPServer(ip string, port int, key string) (*UDPServer, error) {
	block, _ := kcp.NewSalsa20BlockCrypt(sha1.New().Sum([]byte(key)))
	listener, err := kcp.ListenWithOptions(fmt.Sprintf("%s:%d", ip, port), block, 10, 3)
	if err != nil {
		return nil, err
	}
	return &UDPServer{
		addr:     ip,
		port:     port,
		listener: listener,
	}, nil
}

func main() {
	us, err := NewUDPServer("127.0.0.1", 31337, "thisisatest")
	if err != nil {
		log.Fatal(err)
	}
	go us.Start()
	client()
}

func client() {
	key := "thisisatest"
	block, _ := kcp.NewSalsa20BlockCrypt(sha1.New().Sum([]byte(key)))

	// wait for server to become ready
	time.Sleep(time.Second)

	// dial to the echo server
	if sess, err := kcp.DialWithOptions("127.0.0.1:31337", block, 10, 3); err == nil {
		session, _ := smux.Client(sess, nil)
		stream, _ := session.OpenStream()
		for {
			data := time.Now().String()
			data = strings.Repeat(data, 2)
			buf := make([]byte, len(data))
			log.Println("sent:", data)
			if _, err := stream.Write([]byte(data)); err == nil {
				// read back the data
				if _, err := io.ReadFull(stream, buf); err == nil {
					log.Println("recv:", string(buf))
				} else {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
			time.Sleep(time.Second)
		}
	} else {
		log.Fatal(err)
	}
}
