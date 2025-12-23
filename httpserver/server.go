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
	"log"
	"net/http"
	"time"

	liberr "github.com/nabbar/golib/errors"
	loglvl "github.com/nabbar/golib/logger/level"
	libsrv "github.com/nabbar/golib/runner"
)

func (o *srv) getServer() *http.Server {
	if o == nil || o.s == nil {
		return nil
	}

	return o.s.Load()
}

func (o *srv) delServer() {
	if o == nil || o.s == nil {
		return
	}

	o.s.Store(&http.Server{
		ReadHeaderTimeout: time.Nanosecond,
	})
}

func (o *srv) setServer(ctx context.Context) error {
	if o == nil || o.s == nil {
		return ErrorInvalidInstance.Error()
	}

	var (
		ssl  = o.cfgGetTLS()
		bind = o.GetBindable()

		fctStop = func() {
			_ = o.Stop(ctx)
		}
	)

	if o.IsTLS() && ssl == nil {
		err := ErrorServerValidate.Error(fmt.Errorf("TLS Config is not well defined"))
		ent := o.logger().Entry(loglvl.ErrorLevel, "starting http server")
		ent.ErrorAdd(true, err)
		ent.Log()
		return err
	}

	var stdlog = o.logger()

	// #nosec
	s := &http.Server{
		Addr:              bind,
		Handler:           o.HandlerLoadFct(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	stdlog.SetIOWriterFilter("connection reset by peer")

	if ssl != nil && ssl.LenCertificatePair() > 0 {
		s.TLSConfig = ssl.TlsConfig("")
		stdlog.AddIOWriterFilter("TLS handshake error")
	}

	if e := o.cfgGetServer().initServer(s); e != nil {
		ent := o.logger().Entry(loglvl.ErrorLevel, "init http2 server")
		ent.ErrorAdd(true, e)
		ent.Log()
		return e
	}

	s.ErrorLog = stdlog.GetStdLogger(loglvl.ErrorLevel, log.LstdFlags|log.Lmicroseconds)

	if e := o.RunIfPortInUse(ctx, o.GetBindable(), 5, fctStop); e != nil {
		return e
	}

	o.s.Store(s)
	return nil
}

// Start begins serving HTTP requests on the configured bind address.
// The server runs in the background until Stop is called or the context is cancelled.
// Returns an error if the server fails to start or if the port is already in use.
func (o *srv) Start(ctx context.Context) error {
	// Register Runner to runner
	if o.getServer() != nil {
		if e := o.Stop(ctx); e != nil {
			return e
		}
	}

	if e := o.newRun(ctx); e != nil {
		return e
	} else if e = o.runStart(ctx); e != nil {
		return e
	}

	return nil
}

// Stop gracefully shuts down the server without interrupting active connections.
// It waits for active connections to complete up to the context timeout.
// Returns an error if shutdown fails or times out.
func (o *srv) Stop(ctx context.Context) error {
	if o == nil || o.s == nil || o.r == nil {
		return ErrorInvalidInstance.Error()
	}

	r := o.r.Load()
	if r == nil {
		return nil
	}

	return r.Stop(ctx)
}

// Restart performs a stop followed by a start operation.
// This is useful for applying configuration changes that require a server restart.
// Returns an error if either stop or start operations fail.
func (o *srv) Restart(ctx context.Context) error {
	_ = o.Stop(ctx)
	return o.Start(ctx)
}

// IsRunning returns true if the server is currently running and accepting connections.
// Returns false if the server is stopped or has not been started.
func (o *srv) IsRunning() bool {
	if o == nil || o.s == nil || o.r == nil {
		return false
	}

	r := o.r.Load()
	if r == nil {
		return false
	}

	return r.IsRunning()
}

// RunIfPortInUse checks if the specified port is available and retries if in use.
// It calls fct callback if the port is in use, then retries up to nbr times.
// Returns ErrorPortUse if the port remains unavailable after all retries.
func (o *srv) RunIfPortInUse(ctx context.Context, listen string, nbr uint8, fct func()) liberr.Error {
	chk := func() bool {
		return PortInUse(ctx, listen) == nil
	}

	if !libsrv.RunNbr(nbr, chk, fct) {
		return ErrorPortUse.Error(nil)
	}

	return nil
}
