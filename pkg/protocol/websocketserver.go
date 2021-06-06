package protocol

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/dzonerzy/spya/pkg/audio"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type ContextKey string

type WebsocketServer struct {
	srv       *Webserver
	upgrader  websocket.Upgrader
	clients   map[string]*audio.AudioBuffer
	lasterror error
}

func (wss *WebsocketServer) GetLastError() error {
	return wss.lasterror
}

func (wss *WebsocketServer) Stop() {
	wss.lasterror = wss.srv.Stop()
}

func (wss *WebsocketServer) Start(addr string, port int) {
	r := mux.NewRouter()
	r.Handle("/send", wss.addContext(http.HandlerFunc(wss.sendHandler)))
	r.Handle("/recv", wss.addContext(http.HandlerFunc(wss.recvHandler)))
	ws, err := NewWebserver(fmt.Sprintf("%s:%d", addr, port), true, r)
	if err != nil {
		wss.lasterror = err
		return
	}
	wss.srv = ws
	err = wss.srv.Start()
	wss.lasterror = err
}

func (wss *WebsocketServer) Loop() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	wss.Stop()
}

func (wss *WebsocketServer) SetTLS(cert, key string) {
	wss.srv.SetupTLS(cert, key)
}

func (wss *WebsocketServer) createClient(key string) {
	if key != "" {
		if _, ok := wss.clients[key]; !ok {
			wss.clients[key] = audio.NewAudioBuffer(64 * audio.Kilobyte)
		}
	}
}

func (wss *WebsocketServer) removeClient(key string) {
	if key != "" {
		if _, ok := wss.clients[key]; ok {
			wss.clients[key] = nil
			delete(wss.clients, key)
		}
	}
}

func (wss *WebsocketServer) existsClient(key string) bool {
	if key != "" {
		_, ok := wss.clients[key]
		return ok

	}
	return false
}

func (wss *WebsocketServer) sendHandler(rw http.ResponseWriter, r *http.Request) {
	var m = NewMessage()
	wss = r.Context().Value(ContextKey("server")).(*WebsocketServer)
	c, err := wss.upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("Error while upgrading request: %v\n", err)
		return
	}
	log.Printf("Got connection from [%s] receiving audio\n", r.RemoteAddr)
	defer c.Close()
	var key = ""
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			switch err := err.(type) {
			case *websocket.CloseError:
				if err.Code != websocket.CloseNormalClosure {
					log.Printf("Error while reading from websocket: %v\n", err.Error())
				}
			default:
				log.Printf("Error while reading from websocket: %v\n", err)
			}
			wss.removeClient(key)
			return
		}
		// skip first 4 bytes , since that's the packet lenght used only over UDP
		m.Unserialize(message[4:])
		if key == "" {
			if m.Key == "" {
				c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
			key = m.Key
			wss.createClient(key)
		}
		wss.clients[key].WriteSamples(m.Data, int(m.Samples))
	}

}

func (wss *WebsocketServer) recvHandler(rw http.ResponseWriter, r *http.Request) {
	var m = NewMessage()
	log.Printf("Got connection from [%s] reading audio\n", r.RemoteAddr)
	wss = r.Context().Value(ContextKey("server")).(*WebsocketServer)
	c, err := wss.upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("Error while upgrading request: %v\n", err)
		return
	}
	defer c.Close()
	var key = ""
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			switch err := err.(type) {
			case *websocket.CloseError:
				if err.Code != websocket.CloseNormalClosure {
					log.Printf("Error while reading from websocket: %v\n", err.Error())
				}
			default:
				log.Printf("Error while reading from websocket: %v\n", err)
			}
			return
		}
		// skip first 4 bytes , since that's the packet lenght used only over UDP
		m.Unserialize(message[4:])
		if key == "" {
			if m.Key == "" {
				c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
			key = m.Key
			if !wss.existsClient(key) {
				log.Printf("Invalid key received: %s (killing client)\n", key)
				c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
		}
		if m.Samples > 0 {
			if _, ok := wss.clients[key]; !ok {
				return
			}
			data := wss.clients[key].ReadSamples(int(m.Samples))
			m.FromData(data, int(m.Samples))
			m.Type = SendData
			err = c.WriteMessage(websocket.BinaryMessage, m.Serialize())
			if err != nil {
				switch err := err.(type) {
				case *websocket.CloseError:
					if err.Code != websocket.CloseNormalClosure {
						log.Printf("Error while reading from websocket: %v\n", err.Error())
					}
				default:
					log.Printf("Error while reading from websocket: %v\n", err)
				}
				return
			}
		}
	}
}

func (wss *WebsocketServer) addContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ContextKey("server"), wss)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewWebsocketServer() *WebsocketServer {
	wss := &WebsocketServer{
		clients:  make(map[string]*audio.AudioBuffer),
		upgrader: websocket.Upgrader{},
	}
	return wss
}
