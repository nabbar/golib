/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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
	"fmt"
	"runtime"

	logent "github.com/nabbar/golib/logger/entry"
	loglvl "github.com/nabbar/golib/logger/level"
	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

const (
	// DefaultNameMonitor is the default prefix for monitoring identifiers.
	DefaultNameMonitor = "HTTP Server"
)

var (
	// errNotRunning is returned when health check is attempted on a stopped server.
	errNotRunning = errors.New("server is not running")
)

// HealthCheck performs a health check on the server.
// Verifies the server is running, checks for errors, and attempts a TCP connection to the bind address.
// Returns nil if healthy, or an error describing the health issue.
func (o *srv) HealthCheck(ctx context.Context) error {
	var (
		ent logent.Entry
		fl  = func(e ...error) {
			if ent != nil {
				ent.ErrorAdd(true, e...).Check(loglvl.InfoLevel)
			}
		}
	)

	if l := o.logger(); l != nil {
		ent = l.Entry(loglvl.ErrorLevel, "Healthcheck")
	}

	if o.r == nil {
		fl(errNotRunning)
		return errNotRunning
	} else if r := o.r.Load(); r == nil || !r.IsRunning() {
		fl(errNotRunning)
		return errNotRunning
	} else if e := r.ErrorsLast(); e != nil {
		fl(e)
		return e
	} else if e = PortNotUse(ctx, o.GetBindable()); e != nil {
		fl(e)
		return e
	} else {
		fl()
		return nil
	}
}

// MonitorName returns the unique monitoring identifier for this server instance.
// The identifier includes the server's bind address for uniqueness.
func (o *srv) MonitorName() string {
	return fmt.Sprintf("%s [%s]", DefaultNameMonitor, o.GetBindable())
}

// Monitor returns monitoring data for the server including health checks and metrics.
// The vrs parameter provides version information included in the monitoring data.
// Returns a configured Monitor instance or an error if initialization fails.
func (o *srv) Monitor(vrs libver.Version) (montps.Monitor, error) {
	var (
		e   error
		inf moninf.Info
		mon montps.Monitor
		cfg *Config
		res = make(map[string]interface{}, 0)
	)

	if cfg = o.GetConfig(); cfg == nil {
		return nil, fmt.Errorf("cannot load config")
	}

	res["runtime"] = runtime.Version()[2:]
	res["release"] = vrs.GetRelease()
	res["build"] = vrs.GetBuild()
	res["date"] = vrs.GetDate()
	res["handler"] = o.HandlerGetValidKey()

	if inf, e = moninf.New(DefaultNameMonitor); e != nil {
		return nil, e
	} else {
		inf.RegisterName(func() (string, error) {
			return o.MonitorName(), nil
		})
		inf.RegisterInfo(func() (map[string]interface{}, error) {
			return res, nil
		})
	}

	if mon, e = libmon.New(o.c, inf); e != nil {
		return nil, e
	}

	mon.SetHealthCheck(o.HealthCheck)

	if e = mon.SetConfig(o.c, cfg.Monitor); e != nil {
		return nil, e
	}

	if e = mon.Start(o.c.GetContext()); e != nil {
		return nil, e
	}

	return mon, nil
}
