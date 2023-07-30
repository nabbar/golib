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
	"os"
	"sync/atomic"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

type cltu struct {
	a  *atomic.Value // address: hostname + port
	e  *atomic.Value // function error
	i  *atomic.Value // function info
	tr *atomic.Value // connection read timeout
	tw *atomic.Value // connection write timeout
}

func (o *cltu) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.e.Store(f)
}

func (o *cltu) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	o.i.Store(f)
}

func (o *cltu) fctError(e error) {
	if o == nil {
		return
	}

	v := o.e.Load()
	if v != nil {
		v.(libsck.FuncError)(e)
	}
}

func (o *cltu) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	if o == nil {
		return
	}

	v := o.i.Load()
	if v != nil {
		v.(libsck.FuncInfo)(local, remote, state)
	}
}

func (o *cltu) buffRead() *bytes.Buffer {
	return bytes.NewBuffer(make([]byte, 0, 1))
}

func (o *cltu) dial(ctx context.Context) (net.Conn, error) {
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
		return d.DialContext(ctx, libptc.NetworkUDP.Code(), v.(string))
	}
}

func (o *cltu) Do(ctx context.Context, request io.Reader) (io.Reader, error) {
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

	return buf, nil
}
