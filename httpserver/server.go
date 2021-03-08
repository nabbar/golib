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
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nabbar/golib/status"

	liberr "github.com/nabbar/golib/errors"
)

const (
	timeoutShutdown = 10 * time.Second
	timeoutRestart  = 30 * time.Second
)

type server struct {
	run *atomic.Value
	cfg ServerConfig
}

type Server interface {
	GetConfig() *ServerConfig
	SetConfig(cfg *ServerConfig)

	GetName() string
	GetBindable() string
	GetExpose() string
	GetHandlerKey() string

	IsRunning() bool
	IsTLS() bool
	WaitNotify()
	Merge(srv Server) bool

	Listen(handler http.Handler) liberr.Error
	Restart()
	Shutdown()

	StatusInfo() (name string, release string, hash string)
	StatusHealth() error
	StatusComponent(message status.FctMessage) status.Component
}

func NewServer(cfg *ServerConfig) Server {
	return &server{
		cfg: cfg.Clone(),
		run: new(atomic.Value),
	}
}

func (s *server) getRun() run {
	if s.run == nil {
		return newRun()
	} else if i := s.run.Load(); i == nil {
		return newRun()
	} else if r, ok := i.(run); !ok {
		return newRun()
	} else {
		return r
	}
}

func (s *server) setRun(r run) {
	s.run.Store(r)
}

func (s *server) getErr() error {
	if r := s.getRun(); r == nil {
		return nil
	} else {
		return r.GetError()
	}
}

func (s *server) GetConfig() *ServerConfig {
	return &s.cfg
}

func (s *server) SetConfig(cfg *ServerConfig) {
	s.cfg = cfg.Clone()
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

func (s *server) GetHandlerKey() string {
	return s.cfg.GetHandlerKey()
}

func (s *server) IsRunning() bool {
	return s.getRun().IsRunning()
}

func (s *server) IsTLS() bool {
	return s.cfg.IsTLS()
}

func (s *server) Listen(handler http.Handler) liberr.Error {
	r := s.getRun()
	e := r.Listen(&s.cfg, handler)
	s.setRun(r)

	return e
}

func (s *server) WaitNotify() {
	r := s.getRun()
	r.WaitNotify()
}

func (s *server) Restart() {
	_ = s.Listen(nil)
}

func (s *server) Shutdown() {
	r := s.getRun()
	r.Shutdown()
	s.setRun(r)
}

func (s *server) Merge(srv Server) bool {
	if x, ok := srv.(*server); ok {
		s.cfg = x.cfg
		return true
	}

	return false
}

func (s *server) StatusInfo() (name string, release string, hash string) {
	vers := strings.TrimLeft(runtime.Version(), "go")
	vers = strings.TrimLeft(vers, "Go")
	vers = strings.TrimLeft(vers, "GO")

	return fmt.Sprintf("%s [%s]", s.GetName(), s.GetBindable()), vers, ""
}

func (s *server) StatusHealth() error {
	if !s.cfg.Disabled && s.IsRunning() {
		return nil
	} else if s.cfg.Disabled {
		return fmt.Errorf("server disabled")
	} else if e := s.getErr(); e != nil {
		return e
	} else {
		return fmt.Errorf("server is offline -- missing error")
	}
}

func (s *server) StatusComponent(message status.FctMessage) status.Component {
	return status.NewComponent(s.cfg.Mandatory, s.StatusInfo, s.StatusHealth, message, s.cfg.TimeoutCacheInfo, s.cfg.TimeoutCacheHealth)
}
