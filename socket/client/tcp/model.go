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
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"sync/atomic"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

type cltt struct {
	a  *atomic.Value // address: hostname + port
	s  *atomic.Int32 // buffer size
	e  *atomic.Value // function error
	i  *atomic.Value // function info
	tr *atomic.Value // connection read timeout
	tw *atomic.Value // connection write timeout
}

func (o *cltt) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.e.Store(f)
}

func (o *cltt) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	o.i.Store(f)
}

func (o *cltt) fctError(e error) {
	if o == nil {
		return
	}

	v := o.e.Load()
	if v != nil {
		v.(libsck.FuncError)(e)
	}
}

func (o *cltt) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	if o == nil {
		return
	}

	v := o.i.Load()
	if v != nil {
		v.(libsck.FuncInfo)(local, remote, state)
	}
}

func (o *cltt) buffRead() *bytes.Buffer {
	v := o.s.Load()
	if v > 0 {
		return bytes.NewBuffer(make([]byte, 0, int(v)))
	}

	return bytes.NewBuffer(make([]byte, 0, libsck.DefaultBufferSize))
}

func (o *cltt) dial(ctx context.Context) (net.Conn, error) {
	if o == nil {
		return nil, ErrInstance
	}

	v := o.a.Load()
	if v == nil {
		return nil, ErrAddress
	} else if _, e := os.Stat(v.(string)); e != nil {
		return nil, e
	} else {
		d := net.Dialer{}
		return d.DialContext(ctx, libptc.NetworkTCP.Code(), v.(string))
	}
}

func (o *cltt) Do(ctx context.Context, request io.Reader) (io.Reader, error) {
	if o == nil {
		return nil, ErrInstance
	}

	var (
		e error

		lc net.Addr
		rm net.Addr

		cnn net.Conn
		buf = o.buffRead()
	)

	o.fctInfo(nil, nil, libsck.ConnectionDial)
	if cnn, e = o.dial(ctx); e != nil {
		o.fctError(e)
		return nil, e
	}

	defer o.fctError(cnn.Close())

	lc = cnn.LocalAddr()
	rm = cnn.RemoteAddr()

	defer o.fctInfo(lc, rm, libsck.ConnectionClose)
	o.fctInfo(lc, rm, libsck.ConnectionNew)

	if request != nil {
		o.fctInfo(lc, rm, libsck.ConnectionWrite)
		if _, e = io.Copy(cnn, request); e != nil {
			o.fctError(e)
			return nil, e
		}
	}

	o.fctInfo(lc, rm, libsck.ConnectionCloseWrite)
	if e = cnn.(*net.TCPConn).CloseWrite(); e != nil {
		o.fctError(e)
		return nil, e
	}

	o.fctInfo(lc, rm, libsck.ConnectionRead)
	if _, e = io.Copy(buf, cnn); e != nil {
		o.fctError(e)
		return nil, e
	}

	o.fctInfo(lc, rm, libsck.ConnectionCloseRead)
	o.fctError(cnn.(*net.TCPConn).CloseRead())

	return buf, nil
}
