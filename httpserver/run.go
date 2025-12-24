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
	"fmt"
	"net"
	"net/http"
	"time"

	srvtps "github.com/nabbar/golib/httpserver/types"
	loglvl "github.com/nabbar/golib/logger/level"
	libsrv "github.com/nabbar/golib/runner"
	librun "github.com/nabbar/golib/runner/startStop"
)

func (o *srv) newRun(ctx context.Context) error {
	if o == nil || o.r == nil {
		return ErrorServerValidate.Error(nil)
	}

	r := o.r.Swap(librun.New(o.runFuncStart, o.runFuncStop))
	if r != nil && r.IsRunning() {
		if e := r.Stop(ctx); e != nil {
			return e
		}
	}

	return nil
}

func (o *srv) runStart(ctx context.Context) error {
	if o == nil || o.r == nil {
		return ErrorServerValidate.Error(nil)
	}

	r := o.r.Load()
	if r == nil {
		if e := o.newRun(ctx); e != nil {
			return e
		} else if r = o.r.Load(); r == nil {
			return ErrorServerValidate.Error(nil)
		}
	}

	defer func() {
		if r != nil {
			o.r.Store(r)
		}
	}()

	if e := r.Start(ctx); e != nil {
		return e
	}

	var t = time.NewTicker(5 * time.Second)
	defer t.Stop()

	for !r.IsRunning() {
		select {
		case <-ctx.Done():
			return errNotRunning
		default:
			time.Sleep(100 * time.Millisecond)
			if r.IsRunning() {
				return o.GetError()
			}
		}
	}

	return o.GetError()
}

func (o *srv) runFuncStart(ctx context.Context) (err error) {
	var (
		tls = false
		ser *http.Server
	)

	defer func() {
		if r := recover(); r != nil {
			libsrv.RecoveryCaller("golib/httpserver/run/fctStart", r)
		}
		if tls {
			ent := o.logger().Entry(loglvl.InfoLevel, "TLS HTTP Server stopped")
			ent.ErrorAdd(true, err)
			ent.Log()
		} else {
			ent := o.logger().Entry(loglvl.InfoLevel, "HTTP Server stopped")
			ent.ErrorAdd(true, err)
			ent.Log()
		}
	}()

	if ser = o.getServer(); ser == nil {
		if err = o.setServer(ctx); err != nil {
			ent := o.logger().Entry(loglvl.ErrorLevel, "starting http server")
			ent.ErrorAdd(true, err)
			ent.Log()
			return err
		} else if ser = o.getServer(); ser == nil {
			err = ErrorServerStart.Error(fmt.Errorf("cannot create new server, cannot retrieve server"))
			ent := o.logger().Entry(loglvl.ErrorLevel, "starting http server")
			ent.ErrorAdd(true, err)
			ent.Log()
			return err
		}
	}

	if ser.TLSConfig != nil && len(ser.TLSConfig.Certificates) > 0 {
		tls = true
	}

	ser.BaseContext = func(listener net.Listener) context.Context {
		return ctx
	}

	if tls {
		o.logger().Entry(loglvl.InfoLevel, "TLS HTTP Server is starting").Log()
		err = ser.ListenAndServeTLS("", "")
	} else {
		o.logger().Entry(loglvl.InfoLevel, "HTTP Server is starting").Log()
		err = ser.ListenAndServe()
	}

	return err
}

func (o *srv) runFuncStop(ctx context.Context) (err error) {
	var x, n = context.WithTimeout(ctx, srvtps.TimeoutWaitingStop)
	defer n()

	var (
		tls = false
		ser *http.Server
	)

	defer func() {
		if r := recover(); r != nil {
			libsrv.RecoveryCaller("golib/httpserver/run/fctStop", r)
		}
		o.delServer()
		if tls {
			ent := o.logger().Entry(loglvl.InfoLevel, "Shutdown of TLS HTTP Server has been called")
			ent.ErrorAdd(true, err)
			ent.Log()
		} else {
			ent := o.logger().Entry(loglvl.InfoLevel, "Shutdown of HTTP Server has been called")
			ent.ErrorAdd(true, err)
			ent.Log()
		}
	}()

	if ser = o.getServer(); ser == nil {
		err = ErrorServerStart.Error(fmt.Errorf("cannot retrieve server"))
		ent := o.logger().Entry(loglvl.ErrorLevel, "starting http server")
		ent.ErrorAdd(true, err)
		ent.Log()
		return err
	} else if ser.TLSConfig != nil && len(ser.TLSConfig.Certificates) > 0 {
		tls = true
	}

	if tls {
		o.logger().Entry(loglvl.InfoLevel, "Calling TLS HTTP Server shutdown").Log()
	} else {
		o.logger().Entry(loglvl.InfoLevel, "Calling HTTP Server shutdown").Log()
	}

	return ser.Shutdown(x)
}

// Uptime returns the duration since the server was started.
// Returns 0 if the server is not running.
func (o *srv) Uptime() time.Duration {
	if r := o.r.Load(); r != nil {
		return r.Uptime()
	}

	return 0
}

// IsError returns true if the server encountered any errors during operation.
func (o *srv) IsError() bool {
	if r := o.r.Load(); r != nil {
		return r.ErrorsLast() != nil
	}

	return false
}

// GetError returns the last error that occurred during server operation.
// Returns nil if no errors occurred.
func (o *srv) GetError() error {
	if r := o.r.Load(); r != nil {
		return ErrorServerStart.IfError(r.ErrorsList()...)
	}

	return nil
}
