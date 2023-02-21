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
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	liberr "github.com/nabbar/golib/errors"
	srvtps "github.com/nabbar/golib/httpserver/types"
	liblog "github.com/nabbar/golib/logger"
)

func (o *srv) Start(ctx context.Context) error {
	ssl := o.cfgGetTLS()
	if ssl == nil && o.IsTLS() {
		return ErrorServerValidate.ErrorParent(fmt.Errorf("TLS Config is not well defined"))
	}

	bind := o.GetBindable()
	name := o.GetName()
	if name == "" {
		name = bind
	}

	s := &http.Server{
		Addr:     bind,
		ErrorLog: o.logger().GetStdLogger(liblog.ErrorLevel, log.LstdFlags|log.Lmicroseconds),
		Handler:  o.HandlerLoadFct(),
	}

	if ssl != nil && ssl.LenCertificatePair() > 0 {
		s.TLSConfig = ssl.TlsConfig("")
	}

	if e := o.cfgGetServer().initServer(s); e != nil {
		return e
	}

	if o.IsRunning() {
		_ = o.Stop(ctx)
	}

	for i := 0; i < 5; i++ {
		if e := o.PortInUse(o.GetBindable()); e != nil {
			_ = o.Stop(ctx)
		} else {
			break
		}
	}

	o.m.RLock()
	defer o.m.RUnlock()

	o.r.SetServer(s, o.logger, ssl != nil)

	if e := o.r.Start(ctx); e != nil {
		return e
	}

	for i := 0; i < 5; i++ {
		ok := false
		ts := time.Now()

		for {
			time.Sleep(100 * time.Millisecond)
			ok = o.r.IsRunning()

			if ok {
				break
			} else if time.Since(ts) > 10*time.Second {
				break
			}
		}

		if !ok {
			_ = o.r.Restart(ctx)
		} else {
			ts = time.Now()

			for {
				time.Sleep(100 * time.Millisecond)

				if e := o.PortInUse(o.GetBindable()); e != nil {
					return nil
				} else if time.Since(ts) > 10*time.Second {
					break
				}
			}

			_ = o.r.Restart(ctx)
		}
	}

	_ = o.r.Stop(ctx)
	return ErrorServerStart.Error(nil)
}

func (o *srv) Stop(ctx context.Context) error {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.r.Stop(ctx)
}

func (o *srv) Restart(ctx context.Context) error {
	_ = o.Stop(ctx)
	return o.Start(ctx)
}

func (o *srv) IsRunning() bool {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.r.IsRunning()
}

func (o *srv) PortInUse(listen string) liberr.Error {
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

	ctx, cnl = context.WithTimeout(context.TODO(), srvtps.TimeoutWaitingPortFreeing)
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
