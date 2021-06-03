package protocol

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"

	"github.com/dzonerzy/spya/pkg/audio"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type ContextKey string

type AudioServer struct {
	srv      *Webserver
	upgrader websocket.Upgrader
	clients  map[string]*audio.AudioBuffer
}

func (as *AudioServer) Start() (err error) {
	err = as.srv.Start()
	return err
}

func (as *AudioServer) SetTLS(cert, key string) {
	as.srv.SetupTLS(cert, key)
}

func (as *AudioServer) sendHandler(rw http.ResponseWriter, r *http.Request) {
	var secret = r.Header.Get("Client-Secret")
	var m = NewMessage(nil)
	if secret != "" {
		as = r.Context().Value(ContextKey("server")).(*AudioServer)
		c, err := as.upgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Printf("Error while upgrading request: %v\n", err)
			return
		}
		log.Printf("Got connection from [%s] receiving audio\n", r.RemoteAddr)
		as.clients[secret] = audio.NewAudioBuffer(64 * audio.Kilobyte)
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Error while reading from websocket: %v\n", err)
				as.clients[secret] = nil
				return
			}
			m.Unserialize(message)
			as.clients[secret].WriteSamples(m.Data, int(m.Samples))
		}
	}
	log.Printf("Client [%s] sent an empty key\n", r.RemoteAddr)
	rw.WriteHeader(http.StatusNotFound)
}

func (as *AudioServer) recvHandler(rw http.ResponseWriter, r *http.Request) {
	var secret = r.Header.Get("Client-Secret")
	var m = NewMessage(nil)
	if secret != "" {
		log.Printf("Got connection from [%s] reading audio\n", r.RemoteAddr)
		if _, ok := as.clients[secret]; !ok {
			log.Printf("Client [%s] provided a wrong key", r.RemoteAddr)
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		as = r.Context().Value(ContextKey("server")).(*AudioServer)
		c, err := as.upgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Printf("Error while upgrading request: %v\n", err)
			return
		}
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Error while reading from websocket: %v\n", err)
				return
			}
			size := binary.LittleEndian.Uint32(message)
			if size > 0 {
				data := as.clients[secret].ReadSamples(int(size))
				m.FromData(data, int(size))
				err = c.WriteMessage(websocket.BinaryMessage, m.Serialize())
				if err != nil {
					log.Printf("Error while writing from websocket: %v\n", err)
					return
				}
			}
		}
	}
	rw.WriteHeader(http.StatusNotFound)
}

func (as *AudioServer) addContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ContextKey("server"), as)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewAudioServer(addr string, port int) (as *AudioServer, err error) {
	as = &AudioServer{
		clients:  make(map[string]*audio.AudioBuffer),
		upgrader: websocket.Upgrader{},
	}
	r := mux.NewRouter()
	r.Handle("/send", as.addContext(http.HandlerFunc(as.sendHandler)))
	r.Handle("/recv", as.addContext(http.HandlerFunc(as.recvHandler)))
	ws, err := NewWebserver(fmt.Sprintf("%s:%d", addr, port), true, r)
	if err != nil {
		as = nil
		return
	}
	as.srv = ws
	return
}
