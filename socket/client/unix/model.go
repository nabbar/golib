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

package unix

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"sync/atomic"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

type cli struct {
	a *atomic.Value // address : unixfile
	s *atomic.Int32 // buffer size
	e *atomic.Value // function error
	i *atomic.Value // function info
}

func (o *cli) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.e.Store(f)
}

func (o *cli) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	o.i.Store(f)
}

func (o *cli) fctError(e error) {
	if o == nil {
		return
	}

	v := o.e.Load()
	if v != nil {
		v.(libsck.FuncError)(e)
	}
}

func (o *cli) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	if o == nil {
		return
	}

	v := o.i.Load()
	if v != nil {
		v.(libsck.FuncInfo)(local, remote, state)
	}
}

func (o *cli) buffSize() int {
	v := o.s.Load()
	if v > 0 {
		return int(v)
	}

	return libsck.DefaultBufferSize
}

func (o *cli) dial(ctx context.Context) (net.Conn, error) {
	if o == nil {
		return nil, ErrInstance
	}

	v := o.a.Load()

	if v == nil {
		return nil, ErrAddress
	} else if adr, ok := v.(string); !ok {
		return nil, ErrAddress
	} else {
		d := net.Dialer{}
		return d.DialContext(ctx, libptc.NetworkUnix.Code(), adr)
	}
}

func (o *cli) Do(ctx context.Context, request io.Reader, fct libsck.Response) error {
	if o == nil {
		return ErrInstance
	}

	var (
		err error
		cnn net.Conn
	)

	defer func() {
		if cnn != nil {
			o.fctInfo(cnn.LocalAddr(), cnn.RemoteAddr(), libsck.ConnectionClose)
			o.fctError(cnn.Close())
		}
	}()

	o.fctInfo(&net.UnixAddr{}, &net.UnixAddr{}, libsck.ConnectionDial)
	if cnn, err = o.dial(ctx); err != nil {
		o.fctError(err)
		return err
	}

	o.fctInfo(cnn.LocalAddr(), cnn.RemoteAddr(), libsck.ConnectionNew)

	o.sendRequest(cnn, request)
	o.readResponse(cnn, fct)

	return nil
}

func (o *cli) sendRequest(con net.Conn, r io.Reader) {
	var (
		err error
		buf []byte
		rdr = bufio.NewReaderSize(r, o.buffSize())
		wrt = bufio.NewWriterSize(con, o.buffSize())
	)

	defer func() {
		if con != nil {
			if c, ok := con.(*net.UnixConn); ok {
				o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionCloseWrite)
				_ = c.CloseWrite()
			}
		}
	}()

	for {
		if con == nil && r == nil {
			return
		}

		buf, err = rdr.ReadBytes('\n')

		if err != nil {
			if !errors.Is(err, io.EOF) {
				o.fctError(err)
			}
			return
		}

		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionWrite)

		_, err = wrt.Write(buf)
		if err != nil {
			o.fctError(err)
			return
		}

		err = wrt.Flush()
		if err != nil {
			o.fctError(err)
			return
		}
	}
}

func (o *cli) readResponse(con net.Conn, f libsck.Response) {
	if f == nil {
		return
	}

	o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionHandler)
	f(con)
}
