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
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net"
	"sync/atomic"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

func (o *srv) buffSize() int {
	v := o.sr.Load()
	if v > 0 {
		return int(v)
	}

	return libsck.DefaultBufferSize
}

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
		e error
		l net.Listener
		a = o.getAddress()
		s = new(atomic.Bool)
	)

	if len(a) == 0 {
		return ErrInvalidAddress
	} else if hdl := o.handler(); hdl == nil {
		return ErrInvalidHandler
	} else if l, e = o.getListen(a); e != nil {
		o.fctError(e)
		return e
	}

	var fctClose = func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkTCP.String(), a)

		if l != nil {
			_ = l.Close()
		}

		o.r.Store(false)
	}

	defer fctClose()
	s.Store(false)

	go func() {
		<-ctx.Done()
		go func() {
			_ = o.Shutdown()
		}()
		return
	}()

	go func() {
		<-o.Done()

		e = nil
		s.Store(true)

		if l != nil {
			o.fctError(l.Close())
		}

		return
	}()

	o.r.Store(true)
	// Accept new connection or stop if context or shutdown trigger
	for {
		// Accept an incoming connection.
		if l == nil {
			return ErrServerClosed
		} else if s.Load() {
			return e
		}

		if co, ce := l.Accept(); ce != nil && !s.Load() {
			o.fctError(ce)
		} else if co != nil {
			o.fctInfo(co.LocalAddr(), co.RemoteAddr(), libsck.ConnectionNew)
			go o.Conn(co)
		}
	}
}

func (o *srv) Conn(con net.Conn) {
	defer func() {
		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
		_ = con.Close()
	}()

	var (
		err error
		nbr int
		rdr = bufio.NewReaderSize(con, o.buffSize())
		msg []byte
		hdl libsck.Handler
	)

	if hdl = o.handler(); hdl == nil {
		return
	}

	for {
		msg = msg[:0]
		msg, err = rdr.ReadBytes('\n')
		nbr = len(msg)

		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionRead)

		if nbr > 0 {
			if !bytes.HasSuffix(msg, []byte{libsck.EOL}) {
				msg = append(msg, libsck.EOL)
				nbr++
			}

			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionHandler)
			hdl(bytes.NewBuffer(msg[:nbr]), con)
		}

		if err != nil {
			if err != io.EOF {
				o.fctError(err)
			}
			return
		}
	}
}
