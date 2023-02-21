/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package nats

import (
	"context"
	"fmt"
	"runtime"
	"time"

	libctx "github.com/nabbar/golib/context"
	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	natsrv "github.com/nats-io/nats-server/v2/server"
)

const (
	defaultNameMonitor = "Nats Server"
)

func (s *server) HealthCheck(ctx context.Context) error {
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

func (s *server) Monitor(ctx libctx.FuncContext, vrs libver.Version) (montps.Monitor, error) {

	var (
		e   error
		inf moninf.Info
		mon montps.Monitor
		opt *natsrv.Options
		cfg *montps.Config
		res = make(map[string]interface{}, 0)
	)

	s.m.Lock()
	opt = s.GetOptions()
	cfg = s.c
	s.m.Unlock()

	if cfg == nil {
		return nil, fmt.Errorf("cannot load config")
	} else if opt == nil {
		return nil, fmt.Errorf("cannot load config")
	}

	res["runtime"] = runtime.Version()[2:]
	res["release"] = vrs.GetRelease()
	res["build"] = vrs.GetBuild()
	res["date"] = vrs.GetDate()
	res["advertise"] = opt.ClientAdvertise

	if inf, e = moninf.New(defaultNameMonitor); e != nil {
		return nil, e
	} else {
		inf.RegisterName(func() (string, error) {
			return fmt.Sprintf("%s [%s]", defaultNameMonitor, opt.Host), nil
		})
		inf.RegisterInfo(func() (map[string]interface{}, error) {
			return res, nil
		})
	}

	if mon, e = libmon.New(ctx, inf); e != nil {
		return nil, e
	}

	mon.SetHealthCheck(s.HealthCheck)

	if e = mon.SetConfig(ctx, *cfg); e != nil {
		return nil, e
	}

	if e = mon.Start(ctx()); e != nil {
		return nil, e
	}

	return mon, nil
}
