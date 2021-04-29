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
	"strings"
	"sync/atomic"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	natsrv "github.com/nats-io/nats-server/server"
	natcli "github.com/nats-io/nats.go"
)

const (
	DefaultWaitReady = 50 * time.Millisecond
	DefaultTickReady = 5 * DefaultWaitReady
)

type Server interface {
	Listen(ctx context.Context)
	Restart(ctx context.Context)
	Shutdown()

	GetOptions() *natsrv.Options
	SetOptions(opt *natsrv.Options)

	IsRunning() bool
	IsReady() bool
	WaitReady(ctx context.Context, tick time.Duration)

	Client(ctx context.Context, tick time.Duration, defTls libtls.TLSConfig, opt Client) (cli *natcli.Conn, err liberr.Error)

	//StatusInfo() (name string, release string, hash string)
	//StatusHealth() error
	//StatusRoute(prefix string, fctMessage status.FctMessage, sts status.RouteStatus)
}

func NewServer(opt *natsrv.Options) Server {
	o := new(atomic.Value)

	if opt != nil {
		o.Store(opt)
	}

	return &server{
		o: o,
		s: nil,
		r: new(atomic.Value),
	}
}

type server struct {
	o *atomic.Value
	s *natsrv.Server
	r *atomic.Value
}

func (s *server) Listen(ctx context.Context) {
	if s.IsRunning() || s.IsReady() {
		s.Shutdown()
	}

	s.s = natsrv.New(s.GetOptions())
	s.s.ConfigureLogger()
	s.s.Start()
	s.setRunning(true)
	s.WaitReady(ctx, 0)
}

func (s *server) Restart(ctx context.Context) {
	s.Listen(ctx)
}

func (s *server) Shutdown() {
	if s.s != nil {
		s.s.Shutdown()
	}

	s.setRunning(false)
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

func (s *server) setRunning(run bool) {
	if s.r == nil {
		s.r = new(atomic.Value)
	}

	s.r.Store(run)
}

func (s *server) IsReady() bool {
	if s.s != nil {
		return s.s.ReadyForConnections(DefaultWaitReady)
	}

	return false
}

func (s *server) WaitReady(ctx context.Context, tick time.Duration) {
	if tick == 0 {
		tick = DefaultTickReady
	}

	for {
		if s.IsReady() {
			return
		}

		time.Sleep(tick)

		if ctx.Err() != nil {
			return
		}
	}
}

func (s *server) Client(ctx context.Context, tick time.Duration, defTls libtls.TLSConfig, opt Client) (cli *natcli.Conn, err liberr.Error) {
	if opt.Url == "" && len(opt.Servers) == 0 {
		if o := s.GetOptions(); o != nil && o.ClientAdvertise != "" {
			opt.Url = s.formatAddress(o.ClientAdvertise)
		} else if o != nil {
			s.WaitReady(ctx, tick)

			if cAddr := s.s.ClusterAddr(); cAddr != nil && cAddr.String() != "" {
				opt.Url = s.formatAddress(cAddr.String())
				/*
					if o.Cluster.Username != "" {
						opt.User = o.Cluster.Username
					}

					if o.Cluster.Password != "" {
						opt.Password = o.Cluster.Password
					}

				*/
			} else if sAddr := s.s.Addr(); sAddr != nil && sAddr.String() != "" {
				opt.Url = s.formatAddress(sAddr.String())
			} else if o.Host != "" && o.Port > 0 {
				opt.Url = s.formatAddress(fmt.Sprintf("%s:%d", o.Host, o.Port))
			}
		}
	}

	return opt.NewClient(defTls)
}

func (s *server) formatAddress(addr string) string {
	if addr == "" {
		return ""
	}

	if strings.Contains(addr, ",") {
		var b = make([]string, 0)
		for _, a := range strings.Split(addr, ",") {
			b = append(b, s.formatAddress(a))
		}
		return strings.Join(b, ",")
	}

	if strings.HasPrefix(addr, "nats://") {
		return addr
	} else {
		return fmt.Sprintf("nats://%s", addr)
	}
}
