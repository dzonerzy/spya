package protocol

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"

	"github.com/dzonerzy/spya/pkg/ssl"
)

// Webserver define a Webserver
type Webserver struct {
	srv      http.Server
	listener net.Listener
	blocking bool
	handled  []map[string][]interface{}
}

// SetClientTLSAuth add a new client certificate for TLS auth
func (w *Webserver) SetClientTLSAuth(cert string) {
	cacert, err := ssl.PEMToCertificate(cert)
	if err != nil {
		return
	}
	if w.srv.TLSConfig == nil {
		w.srv.TLSConfig = &tls.Config{}
	}
	if w.srv.TLSConfig.ClientCAs == nil {
		w.srv.TLSConfig.ClientCAs = x509.NewCertPool()
	}
	w.srv.TLSConfig.ClientCAs.AddCert(cacert)
	if w.srv.TLSConfig.ClientAuth < tls.RequireAndVerifyClientCert {
		w.srv.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
}

// SetupTLS add TLS support to the Webserver
func (w *Webserver) SetupTLS(cert, key string) {
	certificate, err := ssl.ChainToCertificate(cert, key)
	if err != nil {
		return
	}
	if w.srv.TLSConfig == nil {
		w.srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
	} else {
		w.srv.TLSConfig.Certificates = append(w.srv.TLSConfig.Certificates, certificate)
	}
	w.listener = tls.NewListener(w.listener, w.srv.TLSConfig)
}

// Start the Webserver
func (w *Webserver) Start() error {
	if !w.blocking {
		go func() {
			_ = recover()
			w.srv.Serve(w.listener)
		}()
	} else {
		return w.srv.Serve(w.listener)
	}
	return nil
}

// Stop the Webserver
func (w *Webserver) Stop() error {
	return w.srv.Shutdown(context.Background())
}

// Restart the Webserver
func (w *Webserver) Restart() error {
	err := w.Stop()
	if err != nil {
		return err
	}
	certs := w.srv.TLSConfig.Certificates
	sslauth := w.srv.TLSConfig.ClientAuth
	var pool *x509.CertPool = nil
	if sslauth >= tls.RequestClientCert {
		pool = w.srv.TLSConfig.ClientCAs
	}
	w.srv = http.Server{
		Addr:    w.srv.Addr,
		Handler: &http.ServeMux{},
		//ErrorLog: log.New(ioutil.Discard, "", log.LstdFlags),
	}
	if len(certs) > 0 {
		w.srv.TLSConfig = &tls.Config{
			Certificates: certs,
		}
	}
	if pool != nil {
		w.srv.TLSConfig.ClientCAs = pool
		w.srv.TLSConfig.ClientAuth = sslauth
	}
	w.listener, _ = net.Listen("tcp", w.srv.Addr)
	if w.srv.TLSConfig != nil {
		w.listener = tls.NewListener(w.listener, w.srv.TLSConfig)
	}
	for _, hf := range w.handled {
		for k, v := range hf {
			w.Route(v[0].(string), k, v[1].(http.HandlerFunc))
		}
	}
	err = w.Start()
	if err != nil {
		return err
	}
	return nil
}

func (w *Webserver) Route(virtualhost, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	w.handled = append(w.handled, map[string][]interface{}{pattern: {virtualhost, http.HandlerFunc(handler)}})
	pattern = fmt.Sprintf("%s%s", virtualhost, pattern)
	w.srv.Handler.(*http.ServeMux).HandleFunc(pattern, handler)
}

// NewWebserver return a new Webserver which could be optionally blocking
func NewWebserver(addr string, blocking bool, handler http.Handler) (*Webserver, error) {
	if addr == "" {
		addr = ":http"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	if handler == nil {
		handler = &http.ServeMux{}
	}
	return &Webserver{
		srv: http.Server{
			Addr:    addr,
			Handler: handler,
			//ErrorLog: log.New(ioutil.Discard, "", log.LstdFlags),
		},
		listener: listener,
		blocking: blocking,
	}, nil
}
