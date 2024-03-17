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

package tcp

import (
	"context"
	"crypto/tls"
	"net"
	"sync/atomic"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

func (o *srv) getAddress() string {
	f := o.ad.Load()

	if f != nil {
		return f.(string)
	}

	return ""
}

func (o *srv) getListen(addr string) (net.Listener, error) {
	var (
		lis net.Listener
		err error
	)

	if lis, err = net.Listen(libptc.NetworkTCP.Code(), addr); err != nil {
		return lis, err
	} else if t := o.getTLS(); t != nil {
		lis = tls.NewListener(lis, t)
		o.fctInfoSrv("starting listening socket 'TLS %s %s'", libptc.NetworkTCP.String(), addr)
	} else {
		o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkTCP.String(), addr)
	}

	return lis, nil
}

func (o *srv) Listen(ctx context.Context) error {
	var (
		e error              // error
		l net.Listener       // socket listener
		a = o.getAddress()   // address
		s = new(atomic.Bool) // running
	)

	if len(a) == 0 {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidAddress
	} else if hdl := o.handler(); hdl == nil {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidHandler
	} else if l, e = o.getListen(a); e != nil {
		o.fctError(e)
		return e
	}

	s.Store(false)

	defer func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkTCP.String(), a)

		if l != nil {
			_ = l.Close()
		}

		go func() {
			_ = o.StopGone(context.Background())
		}()

		o.run.Store(false)
	}()

	o.rst.Store(make(chan struct{}))
	o.stp.Store(make(chan struct{}))
	o.run.Store(true)
	o.gon.Store(false)

	go func() {
		defer func() {
			s.Store(true)

			if l != nil {
				o.fctError(l.Close())
			}

			go func() {
				_ = o.Shutdown(context.Background())
			}()
		}()

		select {
		case <-ctx.Done():
			return
		case <-o.Done():
			return
		}
	}()

	// Accept new connection or stop if context or shutdown trigger
	for l != nil && !s.Load() {
		if co, ce := l.Accept(); ce != nil && !s.Load() {
			o.fctError(ce)
		} else if co != nil {
			o.fctInfo(co.LocalAddr(), co.RemoteAddr(), libsck.ConnectionNew)
			go o.Conn(ctx, co)
		}
	}

	return nil
}

func (o *srv) Conn(ctx context.Context, con net.Conn) {
	var (
		hdl libsck.Handler
		cnl context.CancelFunc
		cor libsck.Reader
		cow libsck.Writer
	)

	o.nc.Add(1) // inc nb connection
	ctx, cnl = context.WithCancel(ctx)
	cor, cow = o.getReadWriter(ctx, cnl, con)

	defer func() {
		// cancel context for connection
		cnl()

		// dec nb connection
		o.nc.Add(-1)

		// close connection writer
		_ = cow.Close()

		// delay stopping for 5 seconds to avoid blocking next connection
		if o.IsGone() {
			// if connection is closed
			time.Sleep(5 * time.Second)
		} else {
			// if connection is not closed in 5 seconds
			time.Sleep(500 * time.Millisecond)
		}

		// close connection reader
		_ = cor.Close()

		// send info about connection closing
		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)

		// close connection
		_ = con.Close()
	}()

	// get handler or exit if nil
	if hdl = o.handler(); hdl == nil {
		return
	} else {
		go hdl(cor, cow)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-o.Gone():
			return
		}
	}
}

func (o *srv) getReadWriter(ctx context.Context, cnl context.CancelFunc, con net.Conn) (libsck.Reader, libsck.Writer) {
	var (
		rc = new(atomic.Bool)
		rw = new(atomic.Bool)
	)

	rdrClose := func() error {
		defer func() {
			if rw.Load() {
				cnl()
			}
		}()

		if cr, ok := con.(*net.TCPConn); ok {
			rc.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionCloseRead)
			return cr.CloseRead()
		} else {
			rc.Store(true)
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
			return con.Close()
		}
	}

	wrtClose := func() error {
		defer func() {
			if rc.Load() {
				cnl()
			}
		}()

		if cr, ok := con.(*net.TCPConn); ok {
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionCloseWrite)
			return cr.CloseRead()
		} else {
			rc.Store(true)
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
			return con.Close()
		}
	}

	rdr := libsck.NewReader(
		func(p []byte) (n int, err error) {
			if ctx.Err() != nil {
				_ = rdrClose()
				return 0, ctx.Err()
			}
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionRead)
			return con.Read(p)
		},
		rdrClose,
		func() bool {
			if ctx.Err() != nil {
				_ = rdrClose()
				return false
			}

			_, e := con.Write(nil)

			if e != nil {
				_ = rdrClose()
				return false
			}

			return true
		},
		func() <-chan struct{} {
			return ctx.Done()
		},
	)

	wrt := libsck.NewWriter(
		func(p []byte) (n int, err error) {
			if ctx.Err() != nil {
				_ = wrtClose()
				return 0, ctx.Err()
			}
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionWrite)
			return con.Write(p)
		},
		wrtClose,
		func() bool {
			if ctx.Err() != nil {
				_ = wrtClose()
				return false
			}

			_, e := con.Write(nil)

			if e != nil {
				_ = wrtClose()
				return false
			}

			return true
		},
		func() <-chan struct{} {
			return ctx.Done()
		},
	)

	return rdr, wrt
}
