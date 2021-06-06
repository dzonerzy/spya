package network

import (
	"log"
)

type ClientType int

const (
	ClientSend ClientType = iota + 1
	ClientReceive
)

type ClientData struct {
	Data    []byte
	Samples int
}

type ClientRequest struct {
	Samples int
}

type ClientResponse []byte

type ServerImpl interface {
	Start(addr string, port int)
	Stop()
	Loop()
	GetLastError() error
}

type ClientImpl interface {
	Connect(addr string, port int, clientType ClientType) error
	Disconnect() error
	SendMessage(data []byte, samples int) error
	ReadMessage(samples int) ([]byte, error)
}

func InitClient(clnt ClientImpl, ctype ClientType, addr string, port int, data chan ClientData, request chan ClientRequest, response chan ClientResponse) {
	var err error
	err = clnt.Connect(addr, port, ctype)
	if err != nil {
		log.Fatal(err)
	}
loop:
	for {
		select {
		case d := <-request:
			var data []byte
			data, err = clnt.ReadMessage(d.Samples)
			if err != nil {
				break loop
			}
			response <- data
		case d := <-data:
			err = clnt.SendMessage(d.Data, d.Samples)
			if err != nil {
				break loop
			}
		}
	}
	err = clnt.Disconnect()
	if err != nil {
		log.Fatal(err)
	}
}

func StartServer(srv ServerImpl, addr string, port int, afterInitialization func(srv ServerImpl) error) (err error) {
	go srv.Start(addr, port)
	if afterInitialization != nil {
		err := afterInitialization(srv)
		if err != nil {
			srv.Stop()
			return err
		}
	}
	srv.Loop()
	err = srv.GetLastError()
	return err
}
