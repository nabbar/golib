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

package nats

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	libsts "github.com/nabbar/golib/status"
	natsrv "github.com/nats-io/nats-server/v2/server"
	natcli "github.com/nats-io/nats.go"
)

const (
	DefaultWaitReady = 50 * time.Millisecond
	DefaultTickReady = 5 * DefaultWaitReady
)

type Server interface {
	Listen(ctx context.Context) liberr.Error
	Restart(ctx context.Context) liberr.Error
	Shutdown()

	GetOptions() *natsrv.Options
	SetOptions(opt *natsrv.Options)

	IsRunning() bool
	IsReady() bool
	IsReadyTimeout(parent context.Context, dur time.Duration) bool
	WaitReady(ctx context.Context, tick time.Duration)

	ClientAdvertise(ctx context.Context, tick time.Duration, defTls libtls.TLSConfig, opt Client) (cli *natcli.Conn, err liberr.Error)
	ClientCluster(ctx context.Context, tick time.Duration, defTls libtls.TLSConfig, opt Client) (cli *natcli.Conn, err liberr.Error)
	ClientServer(ctx context.Context, tick time.Duration, defTls libtls.TLSConfig, opt Client) (cli *natcli.Conn, err liberr.Error)

	StatusInfo() (name string, release string, hash string)
	StatusHealth() error
	StatusRouter(sts libsts.RouteStatus, prefix string)
}

func NewServer(opt *natsrv.Options, sts libsts.ConfigStatus) Server {
	o := new(atomic.Value)

	if opt != nil {
		o.Store(opt)
	}

	return &server{
		c: &sts,
		o: o,
		s: nil,
		r: new(atomic.Value),
	}
}

type server struct {
	c *libsts.ConfigStatus
	o *atomic.Value
	s *atomic.Value
	r *atomic.Value
	e liberr.Error
	m sync.Mutex
}

func (s *server) Listen(ctx context.Context) liberr.Error {
	if s.IsRunning() || s.IsReady() {
		s.Shutdown()
	}

	var (
		e error
		o *natsrv.Server
	)

	if o, e = natsrv.NewServer(s.GetOptions()); e != nil {
		err := ErrorServerStart.ErrorParent(e)
		s._SetError(err)
		return err
	}

	s._SetError(nil)
	o.ConfigureLogger()
	o.Start()

	s._SetServer(o)
	s._SetRunning(true)

	//be sure process is launch before trying to check server ready
	time.Sleep(200 * time.Millisecond)
	s.WaitReady(ctx, 0)

	return nil
}

func (s *server) Restart(ctx context.Context) liberr.Error {
	return s.Listen(ctx)
}

func (s *server) Shutdown() {
	if o := s._GetServer(); o != nil {
		o.Shutdown()
	}

	s._SetRunning(false)
}

func (s *server) GetOptions() *natsrv.Options {
	if s.o == nil {
		s.o = new(atomic.Value)
	}

	if i := s.o.Load(); i == nil {
		return nil
	} else if o, ok := i.(*natsrv.Options); !ok {
		return nil
	} else {
		return o
	}
}

func (s *server) SetOptions(opt *natsrv.Options) {
	if opt == nil {
		s.o = new(atomic.Value)
		return
	} else if s.o == nil {
		s.o = new(atomic.Value)
	}

	s.o.Store(opt)
}

func (s *server) IsRunning() bool {
	if s.r == nil {
		s.r = new(atomic.Value)
	}

	if i := s.r.Load(); i == nil {
		return false
	} else if r, ok := i.(bool); !ok {
		return false
	} else {
		return r
	}
}

func (s *server) IsReady() bool {
	if o := s._GetServer(); o != nil {
		return o.ReadyForConnections(DefaultWaitReady)
	}

	return false
}

func (s *server) IsReadyTimeout(parent context.Context, dur time.Duration) bool {
	ctx, cnl := context.WithTimeout(parent, dur)
	defer cnl()

	s.WaitReady(ctx, 0)
	if s.IsRunning() && s.IsReady() {
		return true
	}

	return false
}

func (s *server) WaitReady(ctx context.Context, tick time.Duration) {
	if tick == 0 {
		tick = DefaultTickReady
	}

	for {
		if s.IsRunning() && s.IsReady() {
			return
		}

		time.Sleep(tick)

		if ctx.Err() != nil {
			return
		}
	}
}

func (s *server) ClientAdvertise(ctx context.Context, tick time.Duration, defTls libtls.TLSConfig, opt Client) (cli *natcli.Conn, err liberr.Error) {
	if o := s.GetOptions(); o != nil && o.ClientAdvertise != "" {
		opt.Url = s._FormatAddress(o.ClientAdvertise)
	} else {
		return nil, ErrorConfigValidation.Error(nil)
	}

	return opt.NewClient(defTls)
}

func (s *server) ClientCluster(ctx context.Context, tick time.Duration, defTls libtls.TLSConfig, opt Client) (cli *natcli.Conn, err liberr.Error) {
	s.WaitReady(ctx, tick)

	if srv := s._GetServer(); srv != nil {
		if cAddr := srv.ClusterAddr(); cAddr != nil && cAddr.String() != "" {
			opt.Url = s._FormatAddress(cAddr.String())
		} else {
			return nil, ErrorConfigValidation.Error(nil)
		}
	}

	return opt.NewClient(defTls)
}

func (s *server) ClientServer(ctx context.Context, tick time.Duration, defTls libtls.TLSConfig, opt Client) (cli *natcli.Conn, err liberr.Error) {
	var o *natsrv.Options

	if o = s.GetOptions(); o == nil {
		return nil, ErrorConfigValidation.Error(nil)
	}

	s.WaitReady(ctx, tick)

	if srv := s._GetServer(); srv != nil {
		if sAddr := srv.Addr(); sAddr != nil && sAddr.String() != "" {
			opt.Url = s._FormatAddress(sAddr.String())
		} else if o.Host != "" && o.Port > 0 {
			opt.Url = s._FormatAddress(fmt.Sprintf("%s:%d", o.Host, o.Port))
		} else {
			return nil, ErrorConfigValidation.Error(nil)
		}
	}

	return opt.NewClient(defTls)
}

func (s *server) StatusInfo() (name string, release string, hash string) {
	s.m.Lock()
	defer s.m.Unlock()

	hash = ""
	release = strings.TrimLeft(strings.ToLower(runtime.Version()), "go")
	name = fmt.Sprintf("Nats %d (%s)", s.GetOptions().Host, s.GetOptions().ClientAdvertise)

	return name, release, hash
}

func (s *server) StatusHealth() error {
	for i := 0; i < 5; i++ {
		if s.IsRunning() {
			if s.IsReadyTimeout(context.Background(), time.Second) {
				return nil
			}
		}

		time.Sleep(time.Second)
	}

	if e := s._GetError(); e != nil {
		return e
	}

	return fmt.Errorf("node not ready")
}

func (s *server) StatusRouter(sts libsts.RouteStatus, prefix string) {
	s.m.Lock()
	defer s.m.Unlock()

	if prefix != "" {
		prefix = fmt.Sprintf("%s Nats %d (%s)", prefix, s.GetOptions().Host, s.GetOptions().ClientAdvertise)
	} else {
		prefix = fmt.Sprintf("Nats %d (%s)", s.GetOptions().Host, s.GetOptions().ClientAdvertise)
	}

	s.c.RegisterStatus(sts, prefix, s.StatusInfo, s.StatusHealth)
}

func (s *server) _GetServer() *natsrv.Server {
	s.m.Lock()
	defer s.m.Unlock()

	if s == nil {
		return nil
	} else if s.s == nil {
		s.s = new(atomic.Value)
	}

	if i := s.s.Load(); i == nil {
		return nil
	} else if o, ok := i.(*natsrv.Server); ok {
		return o
	} else {
		return nil
	}
}

func (s *server) _SetServer(srv *natsrv.Server) {
	s.m.Lock()
	defer s.m.Unlock()

	if s == nil {
		return
	} else if s.s == nil {
		s.s = new(atomic.Value)
	}

	s.s.Store(srv)
}

func (s *server) _GetError() liberr.Error {
	s.m.Lock()
	defer s.m.Unlock()

	return s.e
}

func (s *server) _SetError(err liberr.Error) {
	s.m.Lock()
	defer s.m.Unlock()

	s.e = err
}

func (s *server) _SetRunning(run bool) {
	if s.r == nil {
		s.r = new(atomic.Value)
	}

	s.r.Store(run)
}

func (s *server) _FormatAddress(addr string) string {
	if addr == "" {
		return ""
	}

	if strings.Contains(addr, ",") {
		var b = make([]string, 0)
		for _, a := range strings.Split(addr, ",") {
			b = append(b, s._FormatAddress(a))
		}
		return strings.Join(b, ",")
	}

	if strings.HasPrefix(addr, "nats://") {
		return addr
	} else {
		return fmt.Sprintf("nats://%s", addr)
	}
}
