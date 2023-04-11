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

package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	liberr "github.com/nabbar/golib/errors"
	srvtps "github.com/nabbar/golib/httpserver/types"
	liblog "github.com/nabbar/golib/logger"
	libsrv "github.com/nabbar/golib/server"
)

var errInvalid = errors.New("invalid instance")

func (o *srv) getServer() *http.Server {
	if o == nil {
		return nil
	}

	o.m.RLock()
	defer o.m.RUnlock()

	return o.s
}

func (o *srv) delServer() {
	if o == nil {
		return
	}

	o.m.Lock()
	defer o.m.Unlock()

	o.s = nil
}

func (o *srv) setServer(ctx context.Context) error {
	if o == nil {
		return errInvalid
	}

	var (
		ssl  = o.cfgGetTLS()
		bind = o.GetBindable()
		name = o.GetName()

		fctStop = func() {
			_ = o.Stop(ctx)
		}
	)

	if o.IsTLS() && ssl == nil {
		err := ErrorServerValidate.ErrorParent(fmt.Errorf("TLS Config is not well defined"))
		ent := o.logger().Entry(liblog.ErrorLevel, "starting http server")
		ent.ErrorAdd(true, err)
		ent.Log()
		return err
	} else if name == "" {
		name = bind
	}

	var stdlog = o.logger()

	s := &http.Server{
		Addr:    bind,
		Handler: o.HandlerLoadFct(),
	}

	if ssl != nil && ssl.LenCertificatePair() > 0 {
		s.TLSConfig = ssl.TlsConfig("")
		stdlog.SetIOWriterFilter("http: TLS handshake error from 127.0.0.1")
	} else if e := o.cfgGetServer().initServer(s); e != nil {
		ent := o.logger().Entry(liblog.ErrorLevel, "init http2 server")
		ent.ErrorAdd(true, e)
		ent.Log()
		return e
	}

	s.ErrorLog = stdlog.GetStdLogger(liblog.ErrorLevel, log.LstdFlags|log.Lmicroseconds)

	if e := o.RunIfPortInUse(ctx, o.GetBindable(), 5, fctStop); e != nil {
		return e
	}

	o.m.Lock()
	o.s = s
	o.m.Unlock()

	return nil
}

func (o *srv) Start(ctx context.Context) error {
	// Register Server to runner
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

func (o *srv) Stop(ctx context.Context) error {
	if o == nil {
		return errInvalid
	}

	o.m.RLock()
	defer o.m.RUnlock()

	if o.r == nil {
		return nil
	}

	return o.r.Stop(ctx)
}

func (o *srv) Restart(ctx context.Context) error {
	_ = o.Stop(ctx)
	return o.Start(ctx)
}

func (o *srv) IsRunning() bool {
	if o == nil {
		return false
	}

	o.m.RLock()
	defer o.m.RUnlock()

	if o.r == nil {
		return false
	}

	return o.r.IsRunning()
}

func (o *srv) PortInUse(ctx context.Context, listen string) liberr.Error {
	var (
		dia = net.Dialer{}
		con net.Conn
		err error
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

	if strings.Contains(listen, ":") {
		uri := &url.URL{
			Host: listen,
		}

		if h := uri.Hostname(); h == "0.0.0.0" || h == "::1" {
			listen = "127.0.0.1:" + uri.Port()
		}
	}

	if _, ok := ctx.Deadline(); !ok {
		ctx, cnl = context.WithTimeout(ctx, srvtps.TimeoutWaitingPortFreeing)
		defer cnl()
	}

	con, err = dia.DialContext(ctx, "tcp", listen)
	defer func() {
		if con != nil {
			_ = con.Close()
		}
	}()
	if err != nil {
		return nil
	}

	return ErrorPortUse.Error(nil)
}

func (o *srv) PortNotUse(ctx context.Context, listen string) error {
	var (
		err error

		cnl context.CancelFunc
		con net.Conn
		dia = net.Dialer{}
	)

	if strings.Contains(listen, ":") {
		uri := &url.URL{
			Host: listen,
		}

		if h := uri.Hostname(); h == "0.0.0.0" || h == "::1" {
			listen = "127.0.0.1:" + uri.Port()
		}
	}

	if _, ok := ctx.Deadline(); !ok {
		ctx, cnl = context.WithTimeout(ctx, srvtps.TimeoutWaitingPortFreeing)
		defer cnl()
	}

	con, err = dia.DialContext(ctx, "tcp", listen)
	defer func() {
		if con != nil {
			_ = con.Close()
		}
	}()

	return err
}

func (o *srv) RunIfPortInUse(ctx context.Context, listen string, nbr uint8, fct func()) liberr.Error {
	chk := func() bool {
		return o.PortInUse(ctx, listen) == nil
	}

	if !libsrv.RunNbr(nbr, chk, fct) {
		return ErrorPortUse.Error(nil)
	}

	return nil
}
