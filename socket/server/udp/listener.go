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

package udp

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/url"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

func (o *srv) timeoutRead() time.Time {
	v := o.tr.Load()
	if v != nil {
		return time.Now().Add(v.(time.Duration))
	}

	return time.Time{}
}

func (o *srv) timeoutWrite() time.Time {
	v := o.tw.Load()
	if v != nil {
		return time.Now().Add(v.(time.Duration))
	}

	return time.Time{}
}

func (o *srv) buffRead() *bytes.Buffer {
	v := o.sr.Load()
	if v > 0 {
		return bytes.NewBuffer(make([]byte, 0, int(v)))
	}

	return bytes.NewBuffer(make([]byte, 0, libsck.DefaultBufferSize))
}

func (o *srv) getAddress() *url.URL {
	f := o.ad.Load()
	if f != nil {
		return f.(*url.URL)
	}

	return nil
}

func (o *srv) Listen(ctx context.Context) error {
	var (
		e error
		l net.Listener
		a = o.getAddress()
	)

	if a == nil {
		return ErrInvalidAddress
	} else if l, e = net.Listen(libptc.NetworkUDP.Code(), a.Host); e != nil {
		return e
	}

	var fctClose = func() {
		if l != nil {
			o.fctError(l.Close())
		}
	}

	o.fctInfoSrv("starting listening socket 'TLS %s %s'", libptc.NetworkUDP.String(), a.Host)
	defer fctClose()

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
			} else if co, ce := l.Accept(); ce != nil {
				o.fctError(ce)
			} else {
				go o.Conn(co)
			}
		}
	}
}

func (o *srv) Conn(conn net.Conn) {
	defer o.fctError(conn.Close())

	var (
		lc = conn.LocalAddr()
		rm = conn.RemoteAddr()
		tr = o.timeoutRead()
		tw = o.timeoutWrite()
		br = o.buffRead()
	)

	defer o.fctInfo(lc, rm, libsck.ConnectionClose)
	o.fctInfo(lc, rm, libsck.ConnectionNew)

	if !tr.IsZero() {
		if e := conn.SetReadDeadline(tr); e != nil {
			o.fctError(e)
			return
		}
	}

	if !tw.IsZero() {
		if e := conn.SetReadDeadline(tw); e != nil {
			o.fctError(e)
			return
		}
	}

	o.fctInfo(lc, rm, libsck.ConnectionRead)
	if _, e := io.Copy(br, conn); e != nil {
		o.fctError(e)
		return
	}

	if h := o.handler(); h != nil {
		o.fctInfo(lc, rm, libsck.ConnectionHandler)
		h(br, conn)
	}
}
