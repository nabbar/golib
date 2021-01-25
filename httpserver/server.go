/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package httpserver

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/http2"

	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

const (
	timeoutShutdown = 10 * time.Second
	timeoutRestart  = 30 * time.Second
)

type server struct {
	run atomic.Value
	cfg *ServerConfig
	srv *http.Server
	cnl context.CancelFunc
}

type Server interface {
	GetConfig() *ServerConfig
	SetConfig(cfg *ServerConfig)

	GetName() string
	GetBindable() string
	GetExpose() string

	IsRunning() bool
	IsTLS() bool
	WaitNotify()
	Merge(srv Server) bool

	Listen(handler http.Handler) liberr.Error
	Restart()
	Shutdown()
}

func NewServer(cfg *ServerConfig) Server {
	return &server{
		cfg: cfg,
		srv: nil,
		cnl: nil,
	}
}

func (s *server) GetConfig() *ServerConfig {
	return s.cfg
}

func (s *server) SetConfig(cfg *ServerConfig) {
	s.cfg = cfg
}

func (s server) GetName() string {
	if s.cfg.Name == "" {
		s.cfg.Name = s.GetBindable()
	}

	return s.cfg.Name
}

func (s *server) GetBindable() string {
	return s.cfg.GetListen().Host
}

func (s *server) GetExpose() string {
	return s.cfg.GetExpose().String()
}

func (s *server) IsRunning() bool {
	if i := s.run.Load(); i == nil {
		return false
	} else if b, ok := i.(bool); !ok {
		return false
	} else {
		return b
	}
}

func (s *server) IsTLS() bool {
	return s.cfg.IsTLS()
}

func (s *server) setRunning() {
	s.run.Store(true)
}

func (s *server) setNotRunning() {
	s.run.Store(false)
}

func (s *server) Listen(handler http.Handler) liberr.Error {
	ssl, err := s.cfg.GetTLS()
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:     s.GetBindable(),
		ErrorLog: liblog.GetLogger(liblog.ErrorLevel, log.LstdFlags|log.Lmicroseconds, "[http/http2 server '%s']", s.GetName()),
	}

	if s.cfg.ReadTimeout > 0 {
		srv.ReadTimeout = s.cfg.ReadTimeout
	}

	if s.cfg.ReadHeaderTimeout > 0 {
		srv.ReadHeaderTimeout = s.cfg.ReadHeaderTimeout
	}

	if s.cfg.WriteTimeout > 0 {
		srv.WriteTimeout = s.cfg.WriteTimeout
	}

	if s.cfg.MaxHeaderBytes > 0 {
		srv.MaxHeaderBytes = s.cfg.MaxHeaderBytes
	}

	if s.cfg.IdleTimeout > 0 {
		srv.IdleTimeout = s.cfg.IdleTimeout
	}

	if ssl.LenCertificatePair() > 0 {
		srv.TLSConfig = ssl.TlsConfig("")
	}

	if handler != nil {
		srv.Handler = handler
	} else if s.srv != nil {
		srv.Handler = s.srv.Handler
	}

	cfg := &http2.Server{}

	if s.cfg.MaxHandlers > 0 {
		cfg.MaxHandlers = s.cfg.MaxHandlers
	}

	if s.cfg.MaxConcurrentStreams > 0 {
		cfg.MaxConcurrentStreams = s.cfg.MaxConcurrentStreams
	}

	if s.cfg.PermitProhibitedCipherSuites {
		cfg.PermitProhibitedCipherSuites = true
	}

	if s.cfg.IdleTimeout > 0 {
		cfg.IdleTimeout = s.cfg.IdleTimeout
	}

	if s.cfg.MaxUploadBufferPerConnection > 0 {
		cfg.MaxUploadBufferPerConnection = s.cfg.MaxUploadBufferPerConnection
	}

	if s.cfg.MaxUploadBufferPerStream > 0 {
		cfg.MaxUploadBufferPerStream = s.cfg.MaxUploadBufferPerStream
	}

	if e := http2.ConfigureServer(srv, cfg); e != nil {
		return ErrorHTTP2Configure.ErrorParent(e)
	}

	if s.IsRunning() {
		s.Shutdown()
	}

	for i := 0; i < 5; i++ {
		if e := s.PortInUse(); e != nil {
			s.Shutdown()
		} else {
			break
		}
	}

	s.srv = srv

	go func() {
		ctx, cnl := context.WithCancel(s.cfg.getContext())
		s.cnl = cnl

		defer func() {
			cnl()
			s.setNotRunning()
		}()

		s.srv.BaseContext = func(listener net.Listener) context.Context {
			return ctx
		}

		var err error

		if ssl.LenCertificatePair() > 0 {
			liblog.InfoLevel.Logf("TLS Server '%s' is starting with bindable: %s", s.GetName(), s.GetBindable())

			s.setRunning()
			err = s.srv.ListenAndServeTLS("", "")
		} else {
			liblog.InfoLevel.Logf("Server '%s' is starting with bindable: %s", s.GetName(), s.GetBindable())

			s.setRunning()
			err = s.srv.ListenAndServe()
		}

		if err != nil && ctx.Err() != nil && ctx.Err().Error() == err.Error() {
			return
		} else if err != nil && errors.Is(err, http.ErrServerClosed) {
			return
		} else if err != nil {
			liblog.ErrorLevel.LogErrorCtxf(liblog.NilLevel, "Listen Server '%s'", err, s.GetName())
		}
	}()

	return nil
}

func (s *server) WaitNotify() {
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	signal.Notify(quit, syscall.SIGQUIT)

	select {
	case <-quit:
		s.Shutdown()
	case <-s.cfg.getContext().Done():
		s.Shutdown()
	}
}

func (s *server) Restart() {
	_ = s.Listen(nil)
}

func (s *server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutShutdown)
	defer func() {
		cancel()

		if s.srv != nil {
			_ = s.srv.Close()
		}

		s.setNotRunning()
	}()

	liblog.InfoLevel.Logf("Shutdown Server '%s'...", s.GetName())

	if s.cnl != nil {
		s.cnl()
	}

	if s.srv != nil {
		err := s.srv.Shutdown(ctx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			liblog.ErrorLevel.Logf("Shutdown Server '%s' Error: %v", s.GetName(), err)
		}
	}
}

func (s *server) Merge(srv Server) bool {
	if x, ok := srv.(*server); ok {
		s.cfg = x.cfg
		return true
	}

	return false
}

func (s *server) PortInUse() liberr.Error {
	var (
		dia = net.Dialer{}
		con net.Conn
		err error
		ctx context.Context
		cnl context.CancelFunc
	)

	defer func() {
		if cnl != nil {
			cnl()
		}
		if con != nil {
			_ = con.Close()
		}
	}()

	ctx, cnl = context.WithTimeout(context.TODO(), 2*time.Second)
	con, err = dia.DialContext(ctx, "tcp", s.cfg.Listen)

	if err != nil {
		return nil
	}

	return ErrorPortUse.Error(nil)
}
