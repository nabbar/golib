/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	"golang.org/x/net/http2"
)

type srvRun struct {
	err *atomic.Value
	run *atomic.Value
	snm string
	srv *http.Server
	ctx context.Context
	cnl context.CancelFunc
}

type run interface {
	IsRunning() bool
	GetError() error
	WaitNotify()
	Listen(cfg *ServerConfig, handler http.Handler) liberr.Error
	Restart(cfg *ServerConfig)
	Shutdown()
}

func newRun() run {
	return &srvRun{
		err: new(atomic.Value),
		run: new(atomic.Value),
		srv: nil,
	}
}

func (s *srvRun) getRunning() bool {
	if s.run == nil {
		return false
	} else if i := s.run.Load(); i == nil {
		return false
	} else if b, ok := i.(bool); !ok {
		return false
	} else {
		return b
	}
}

func (s *srvRun) setRunning(state bool) {
	s.run.Store(state)
}

func (s *srvRun) IsRunning() bool {
	return s.getRunning()
}

func (s *srvRun) getErr() error {
	if s.err == nil {
		return nil
	} else if i := s.err.Load(); i == nil {
		return nil
	} else if e, ok := i.(error); !ok {
		return nil
	} else if e.Error() == "" {
		return nil
	} else {
		return e
	}
}

func (s *srvRun) setErr(e error) {
	if e != nil {
		s.err.Store(e)
	} else {
		s.err.Store(errors.New(""))
	}
}

func (s *srvRun) GetError() error {
	return s.getErr()
}

func (s *srvRun) WaitNotify() {
	if !s.IsRunning() {
		return
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	signal.Notify(quit, syscall.SIGQUIT)

	select {
	case <-quit:
		s.Shutdown()
	case <-s.ctx.Done():
		s.Shutdown()
	}
}

func (s *srvRun) Merge(srv Server) bool {
	panic("implement me")
}

func (s *srvRun) Listen(cfg *ServerConfig, handler http.Handler) liberr.Error {
	ssl, err := cfg.GetTLS()
	if err != nil {
		return err
	}

	sTls := cfg.TLSMandatory
	bind := cfg.GetListen().Host
	name := cfg.Name
	if name == "" {
		name = bind
	}

	srv := &http.Server{
		Addr:     cfg.GetListen().Host,
		ErrorLog: liblog.GetLogger(liblog.ErrorLevel, log.LstdFlags|log.Lmicroseconds, "[http/http2 server '%s']", name),
	}

	if cfg.ReadTimeout > 0 {
		srv.ReadTimeout = cfg.ReadTimeout
	}

	if cfg.ReadHeaderTimeout > 0 {
		srv.ReadHeaderTimeout = cfg.ReadHeaderTimeout
	}

	if cfg.WriteTimeout > 0 {
		srv.WriteTimeout = cfg.WriteTimeout
	}

	if cfg.MaxHeaderBytes > 0 {
		srv.MaxHeaderBytes = cfg.MaxHeaderBytes
	}

	if cfg.IdleTimeout > 0 {
		srv.IdleTimeout = cfg.IdleTimeout
	}

	if ssl.LenCertificatePair() > 0 {
		srv.TLSConfig = ssl.TlsConfig("")
	}

	if handler != nil {
		srv.Handler = handler
	} else if s.srv != nil {
		srv.Handler = s.srv.Handler
	}

	s2 := &http2.Server{}

	if cfg.MaxHandlers > 0 {
		s2.MaxHandlers = cfg.MaxHandlers
	}

	if cfg.MaxConcurrentStreams > 0 {
		s2.MaxConcurrentStreams = cfg.MaxConcurrentStreams
	}

	if cfg.PermitProhibitedCipherSuites {
		s2.PermitProhibitedCipherSuites = true
	}

	if cfg.IdleTimeout > 0 {
		s2.IdleTimeout = cfg.IdleTimeout
	}

	if cfg.MaxUploadBufferPerConnection > 0 {
		s2.MaxUploadBufferPerConnection = cfg.MaxUploadBufferPerConnection
	}

	if cfg.MaxUploadBufferPerStream > 0 {
		s2.MaxUploadBufferPerStream = cfg.MaxUploadBufferPerStream
	}

	if e := http2.ConfigureServer(srv, s2); e != nil {
		s.setErr(e)
		return ErrorHTTP2Configure.ErrorParent(e)
	}

	if s.IsRunning() {
		s.Shutdown()
	}

	for i := 0; i < 5; i++ {
		if e := s.PortInUse(cfg.Listen); e != nil {
			s.Shutdown()
		} else {
			break
		}
	}

	if s.ctx != nil && s.ctx.Err() == nil && s.cnl != nil {
		s.cnl()
		s.ctx = nil
		s.cnl = nil
	}

	s.ctx, s.cnl = context.WithCancel(cfg.getContext())
	s.snm = name
	s.srv = srv

	go func(name, host string, tlsMandatory bool) {

		defer func() {
			if s.ctx != nil && s.cnl != nil && s.ctx.Err() == nil {
				s.cnl()
			}
			s.setRunning(false)
		}()

		s.srv.BaseContext = func(listener net.Listener) context.Context {
			return s.ctx
		}

		var err error

		if ssl.LenCertificatePair() > 0 {
			liblog.InfoLevel.Logf("TLS Server '%s' is starting with bindable: %s", name, host)

			s.setRunning(true)
			err = s.srv.ListenAndServeTLS("", "")
		} else if tlsMandatory {
			err = fmt.Errorf("missing valid server certificates")
		} else {
			liblog.InfoLevel.Logf("Server '%s' is starting with bindable: %s", name, host)

			s.setRunning(true)
			err = s.srv.ListenAndServe()
		}

		if err != nil && s.ctx.Err() != nil && s.ctx.Err().Error() == err.Error() {
			return
		} else if err != nil && errors.Is(err, http.ErrServerClosed) {
			return
		} else if err != nil {
			s.setErr(err)
			liblog.ErrorLevel.LogErrorCtxf(liblog.NilLevel, "Listen Server '%s'", err, name)
		}
	}(name, bind, *sTls)

	return nil
}

func (s *srvRun) Restart(cfg *ServerConfig) {
	_ = s.Listen(cfg, nil)
}

func (s *srvRun) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutShutdown)

	defer func() {
		cancel()

		if s.srv != nil {
			_ = s.srv.Close()
		}

		s.setRunning(false)
	}()

	liblog.InfoLevel.Logf("Shutdown Server '%s'...", s.snm)

	if s.cnl != nil && s.ctx != nil && s.ctx.Err() == nil {
		s.cnl()
	}

	if s.srv != nil {
		err := s.srv.Shutdown(ctx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			liblog.ErrorLevel.Logf("Shutdown Server '%s' Error: %v", s.snm, err)
		}
	}
}

func (s *srvRun) PortInUse(listen string) liberr.Error {
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
	con, err = dia.DialContext(ctx, "tcp", listen)

	if con != nil {
		_ = con.Close()
		con = nil
	}

	cnl()
	cnl = nil

	if err != nil {
		return nil
	}

	return ErrorPortUse.Error(nil)
}
