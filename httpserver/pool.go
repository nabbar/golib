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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	libsem "github.com/nabbar/golib/semaphore"
	libsts "github.com/nabbar/golib/status"
)

type FieldType uint8

const (
	HandlerDefault           = "default"
	FieldName      FieldType = iota
	FieldBind
	FieldExpose
)

type MapUpdPoolServer func(srv Server) Server
type MapRunPoolServer func(srv Server)

type pool []Server

type PoolServer interface {
	Add(srv ...Server) (PoolServer, liberr.Error)
	Get(bindAddress string) Server
	Del(bindAddress string) PoolServer
	Has(bindAddress string) bool
	Len() int

	MapRun(f MapRunPoolServer)
	MapUpd(f MapUpdPoolServer)

	List(fieldFilter, fieldReturn FieldType, pattern, regex string) []string
	Filter(field FieldType, pattern, regex string) PoolServer

	IsRunning(asLeast bool) bool
	WaitNotify(ctx context.Context, cancel context.CancelFunc)

	Listen(handler http.Handler) liberr.Error
	ListenMultiHandler(handler map[string]http.Handler) liberr.Error
	Restart()
	Shutdown()

	StatusInfo(bindAddress string) (name string, release string, hash string)
	StatusHealth(bindAddress string) error
	StatusRoute(prefix string, fctMessage libsts.FctMessage, sts libsts.RouteStatus)
}

func NewPool(srv ...Server) PoolServer {
	p, _ := make(pool, 0).Add(srv...)
	return p
}

func (p pool) MapRun(f MapRunPoolServer) {
	if p == nil {
		return
	}

	for _, s := range p {
		f(s)
	}
}

func (p pool) MapUpd(f MapUpdPoolServer) {
	if p == nil {
		return
	}

	for i, s := range p {
		p[i] = f(s)
	}
}

func (p pool) Add(srv ...Server) (PoolServer, liberr.Error) {
	var r = make(pool, 0)

	if p != nil {
		r = p
	}

	for _, s := range srv {
		if !r.Has(s.GetBindable()) {
			r = append(r, s)
			continue
		}

		for _, x := range r {
			if x.GetBindable() != s.GetBindable() {
				continue
			} else if !x.Merge(s) {
				r = r.Del(s.GetBindable()).(pool)
				r = append(r, s)
				break
			}
		}
	}

	return r, nil
}

func (p pool) Get(bindAddress string) Server {
	if !p.Has(bindAddress) {
		return nil
	}

	for _, s := range p {
		if s.GetBindable() == bindAddress {
			return s
		}
	}

	return nil
}

func (p pool) Del(bindAddress string) PoolServer {
	if !p.Has(bindAddress) {
		return p
	}

	var r = make(pool, 0)

	for _, s := range p {
		if s.GetBindable() != bindAddress {
			r = append(r, s)
		}

		if s.IsRunning() {
			s.Shutdown()
		}
	}

	return r
}

func (p pool) Has(bindAddress string) bool {
	if p.Len() < 1 {
		return false
	}

	for _, s := range p {
		if s.GetBindable() == bindAddress {
			return true
		}
	}

	return false
}

func (p pool) Len() int {
	if p == nil {
		return 0
	}

	return len(p)
}

func (p pool) List(fieldFilter, fieldReturn FieldType, pattern, regex string) []string {
	var (
		r = make([]string, 0)
		f string
	)

	if p.Len() < 1 {
		return r
	}

	pattern = strings.ToLower(pattern)

	p.MapRun(func(srv Server) {
		switch fieldFilter {
		case FieldBind:
			f = srv.GetBindable()
		case FieldExpose:
			f = srv.GetExpose()
		case FieldName:
			f = srv.GetName()
		default:
			f = srv.GetName()
		}

		f = strings.ToLower(f)

		if pattern != "" && strings.Contains(f, pattern) {
			switch fieldReturn {
			case FieldBind:
				r = append(r, srv.GetBindable())
			case FieldExpose:
				r = append(r, srv.GetExpose())
			case FieldName:
				r = append(r, srv.GetName())
			default:
				r = append(r, srv.GetName())
			}

			return
		}

		if regex == "" {
			return
		}

		if ok, err := regexp.MatchString(regex, srv.GetName()); err == nil && ok {
			switch fieldReturn {
			case FieldBind:
				r = append(r, srv.GetBindable())
			case FieldExpose:
				r = append(r, srv.GetExpose())
			case FieldName:
				r = append(r, srv.GetName())
			default:
				r = append(r, srv.GetName())
			}
		}
	})

	return r
}

func (p pool) Filter(field FieldType, pattern, regex string) PoolServer {
	if p.Len() < 1 {
		return nil
	}

	var (
		r = make(pool, 0)
		f string
	)

	pattern = strings.ToLower(pattern)

	p.MapRun(func(srv Server) {
		switch field {
		case FieldBind:
			f = srv.GetBindable()
		case FieldExpose:
			f = srv.GetExpose()
		case FieldName:
			f = srv.GetName()
		default:
			f = srv.GetName()
		}

		f = strings.ToLower(f)

		if pattern != "" && strings.Contains(f, pattern) {
			r = append(r, srv)
			return
		}

		if regex == "" {
			return
		}

		if ok, err := regexp.MatchString(regex, srv.GetName()); err == nil && ok {
			r = append(r, srv)
		}
	})

	return r
}

func (p pool) IsRunning(atLeast bool) bool {
	if p.Len() < 1 {
		return false
	}

	var r = false

	for _, s := range p {
		if s.IsRunning() {
			r = true
			continue
		}

		if !atLeast {
			return false
		}
	}

	return r
}

func (p pool) WaitNotify(ctx context.Context, cancel context.CancelFunc) {
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	signal.Notify(quit, syscall.SIGQUIT)

	select {
	case <-quit:
		p.Shutdown()
		if cancel != nil {
			cancel()
		}
	case <-ctx.Done():
		p.Shutdown()
		if cancel != nil {
			cancel()
		}
	}
}

func (p pool) Listen(handler http.Handler) liberr.Error {
	return p.ListenMultiHandler(map[string]http.Handler{HandlerDefault: handler})
}

func (p pool) ListenMultiHandler(handler map[string]http.Handler) liberr.Error {
	if p.Len() < 1 {
		return nil
	}

	var e liberr.Error

	e = ErrorPoolListen.Error(nil)
	liblog.InfoLevel.Log("Calling listen for All Servers")

	p.MapRun(func(srv Server) {
		if len(handler) < 1 {
			e.AddParentError(srv.Listen(nil))
		} else {
			for k := range handler {
				if len(handler) == 1 {
					e.AddParentError(srv.Listen(handler[k]))
					break
				} else if strings.ToLower(k) == srv.GetHandlerKey() {
					e.AddParentError(srv.Listen(handler[k]))
					break
				}
			}
		}
	})

	liblog.InfoLevel.Log("End of Calling listen for All Servers")

	if !e.HasParent() {
		e = nil
	}

	return e
}

func (p pool) runMapCommand(f func(sem libsem.Sem, srv Server)) {
	if p.Len() < 1 {
		return
	}

	var (
		s libsem.Sem
		x context.Context
		c context.CancelFunc
	)

	x, c = context.WithTimeout(context.Background(), timeoutShutdown)

	defer func() {
		c()
		s.DeferMain()
	}()

	s = libsem.NewSemaphoreWithContext(x, 0)

	p.MapRun(func(srv Server) {
		_ = s.NewWorker()
		go func(sem libsem.Sem, srv Server) {
			f(sem, srv)
		}(s, srv)
	})

	_ = s.WaitAll()
}

func (p pool) runMapRestart(sem libsem.Sem, srv Server) {
	defer func() {
		if sem != nil {
			sem.DeferWorker()
		}
	}()

	if srv != nil {
		srv.Restart()
	}
}

func (p pool) runMapShutdown(sem libsem.Sem, srv Server) {
	defer func() {
		if sem != nil {
			sem.DeferWorker()
		}
	}()

	if srv != nil {
		srv.Shutdown()
	}
}

func (p pool) Restart() {
	p.runMapCommand(p.runMapRestart)
}

func (p pool) Shutdown() {
	p.runMapCommand(p.runMapShutdown)
}

func (p pool) StatusInfo(bindAddress string) (name string, release string, hash string) {
	if s := p.Get(bindAddress); s != nil {
		return s.StatusInfo()
	}

	return fmt.Sprintf("missing server '%s'", bindAddress), "", ""
}

func (p pool) StatusHealth(bindAddress string) error {
	if s := p.Get(bindAddress); s != nil {
		return s.StatusHealth()
	}

	//nolint #goerr113
	return fmt.Errorf("missing server '%s'", bindAddress)
}

func (p pool) StatusRoute(keyPrefix string, fctMessage libsts.FctMessage, sts libsts.RouteStatus) {
	p.MapRun(func(srv Server) {
		bind := srv.GetBindable()
		sts.ComponentNew(fmt.Sprintf("%s-%s", keyPrefix, bind), srv.StatusComponent(fctMessage))
	})
}
