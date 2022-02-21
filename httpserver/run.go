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
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	"golang.org/x/net/http2"
)

const _TimeoutWaitingPortFreeing = 500 * time.Microsecond

type srvRun struct {
	m sync.Mutex

	log liblog.FuncLog     // return golib logger interface
	err *atomic.Value      // last err occured
	run *atomic.Value      // is running
	snm string             // server name
	bnd string             // server bind
	tls bool               // is tls server or tls mandatory
	srv *http.Server       // golang http server
	ctx context.Context    // context
	cnl context.CancelFunc // cancel func of context
}

type run interface {
	IsRunning() bool
	GetError() error
	WaitNotify()
	Listen(cfg *ServerConfig, handler http.Handler) liberr.Error
	Restart(cfg *ServerConfig)
	Shutdown()
}

func newRun(log liblog.FuncLog) run {
	return &srvRun{
		m:   sync.Mutex{},
		log: log,
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
		//nolint #goerr113
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
		if s.IsRunning() {
			s.Shutdown()
		} else if s.srv != nil {
			s.srvShutdown()
		}
	}
}

func (s *srvRun) Merge(srv Server) bool {
	panic("implement me")
}

func (s *srvRun) getLogger() liblog.Logger {
	var _log liblog.Logger

	if s.log == nil {
		_log = liblog.GetDefault()
	} else if l := s.log(); l == nil {
		_log = liblog.GetDefault()
	} else {
		_log = l
	}

	_log.SetFields(_log.GetFields().Add("server_name", s.snm).Add("server_bind", s.bnd).Add("server_tls", s.tls))
	return _log
}

//nolint #gocognit
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

	s.snm = name
	s.bnd = bind
	s.tls = sTls || ssl.LenCertificatePair() > 0

	srv := &http.Server{
		Addr:     cfg.GetListen().Host,
		ErrorLog: s.getLogger().GetStdLogger(liblog.ErrorLevel, log.LstdFlags|log.Lmicroseconds),
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

	s.m.Lock()
	s.srv = srv
	s.m.Unlock()

	go func(ctx context.Context, cnl context.CancelFunc, name, host string, tlsMandatory bool) {
		var _log = s.getLogger()
		ent := _log.Entry(liblog.InfoLevel, "server stopped")

		defer func() {
			ent.Log()
			if ctx != nil && cnl != nil && ctx.Err() == nil {
				cnl()
			}
			s.setRunning(false)
		}()

		s.m.Lock()
		if s.srv == nil {
			return
		}
		s.srv.BaseContext = func(listener net.Listener) context.Context {
			return s.ctx
		}
		s.m.Unlock()

		var er error
		_log.Entry(liblog.InfoLevel, "Server is starting").Log()

		if ssl.LenCertificatePair() > 0 {
			s.setRunning(true)
			er = s.srv.ListenAndServeTLS("", "")
		} else if tlsMandatory {
			//nolint #goerr113
			er = fmt.Errorf("missing valid server certificates")
		} else {
			s.setRunning(true)
			er = s.srv.ListenAndServe()
		}

		if er != nil && ctx.Err() != nil && ctx.Err().Error() == er.Error() {
			return
		} else if er != nil && errors.Is(er, http.ErrServerClosed) {
			return
		} else if er != nil {
			s.setErr(er)
			ent.Level = liblog.ErrorLevel
			ent.ErrorAdd(true, er)
		}
	}(s.ctx, s.cnl, name, bind, sTls)

	return nil
}

func (s *srvRun) Restart(cfg *ServerConfig) {
	_ = s.Listen(cfg, nil)
}

func (s *srvRun) Shutdown() {
	s.srvShutdown()
	s.setRunning(false)

	if s.cnl != nil && s.ctx != nil && s.ctx.Err() == nil {
		s.cnl()
	}
}

func (s *srvRun) srvShutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutShutdown)
	_log := s.getLogger()

	defer func() {
		cancel()
		if s.srv != nil {
			err := s.srv.Close()

			s.m.Lock()
			s.srv = nil
			s.m.Unlock()

			_log.Entry(liblog.ErrorLevel, "closing server").ErrorAdd(true, err).Check(liblog.InfoLevel)
		}
	}()

	if s.srv != nil {
		err := s.srv.Shutdown(ctx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			_log.Entry(liblog.ErrorLevel, "Shutdown server").ErrorAdd(true, err).Check(liblog.InfoLevel)
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

	ctx, cnl = context.WithTimeout(context.TODO(), _TimeoutWaitingPortFreeing)
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
