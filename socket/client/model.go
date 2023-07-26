//go:build linux
// +build linux

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

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync/atomic"

	libsck "github.com/nabbar/golib/socket"
)

type clt struct {
	u  *atomic.Value // unixfile
	f  *atomic.Value // function error
	tr *atomic.Value // connection read timeout
	tw *atomic.Value // connection write timeout
}

func (o *clt) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.f.Store(f)
}

func (o *clt) fctError(e error) {
	if o == nil {
		return
	}

	v := o.f.Load()
	if v != nil {
		v.(libsck.FuncError)(e)
	}
}

func (o *clt) dial(ctx context.Context) (net.Conn, error) {
	if o == nil {
		return nil, fmt.Errorf("invalid instance")
	}

	v := o.u.Load()
	if v == nil {
		return nil, fmt.Errorf("invalid unix file")
	} else if _, e := os.Stat(v.(string)); e != nil {
		return nil, e
	} else {
		d := net.Dialer{}
		return d.DialContext(ctx, "unix", v.(string))
	}
}

func (o *clt) Connection(ctx context.Context, request io.Reader) (io.Reader, error) {
	if o == nil {
		return nil, fmt.Errorf("invalid instance")
	}

	var (
		e error

		cnn net.Conn
	)

	if cnn, e = o.dial(ctx); e != nil {
		o.fctError(e)
		return nil, e
	}

	defer cnn.Close()

	if request != nil {
		if _, e = io.Copy(cnn, request); e != nil {
			o.fctError(e)
			return nil, e
		}
	}

	if e = cnn.(*net.UnixConn).CloseWrite(); e != nil {
		o.fctError(e)
		return nil, e
	}

	var buf = bytes.NewBuffer(make([]byte, 0, 32*1024))

	if _, e = io.Copy(buf, cnn); e != nil {
		o.fctError(e)
		return nil, e
	}

	_ = cnn.(*net.UnixConn).CloseRead()

	return buf, nil
}
