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
	)

	var fctClose = func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkTCP.String(), a)
		if l != nil {
			o.fctError(l.Close())
		}
	}

	defer fctClose()

	if len(a) == 0 {
		return ErrInvalidAddress
	} else if l, e = o.getListen(a); e != nil {
		o.fctError(e)
		return e
	}

	// Accept new connection or stop if context or shutdown trigger
	for {
		select {
		case <-ctx.Done():
			return ErrContextClosed
		case <-o.Done():
			return nil
		default:
			// Accept an incoming connection.
			if l == nil {
				return ErrServerClosed
			}

			co, ce := l.Accept()

			if ce != nil {
				o.fctError(ce)
			} else {
				go o.Conn(co)
			}
		}
	}
}

func (o *srv) Conn(conn net.Conn) {
	defer func() {
		o.fctInfo(conn.LocalAddr(), conn.RemoteAddr(), libsck.ConnectionClose)
		_ = conn.Close()
	}()

	o.fctInfo(conn.LocalAddr(), conn.RemoteAddr(), libsck.ConnectionNew)

	var (
		err error
		rdr = bufio.NewReaderSize(conn, o.buffSize())
		buf []byte
		hdl libsck.Handler
	)

	if hdl = o.handler(); hdl == nil {
		return
	}

	for {
		buf, err = rdr.ReadBytes('\n')

		o.fctInfo(conn.LocalAddr(), conn.RemoteAddr(), libsck.ConnectionRead)
		if err != nil {
			if err != io.EOF {
				o.fctError(err)
			}
			break
		}

		o.fctInfo(conn.LocalAddr(), conn.RemoteAddr(), libsck.ConnectionHandler)
		hdl(bytes.NewBuffer(buf), conn)
	}
}
