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
	"context"
	"errors"
	"io"
	"net"
	"sync/atomic"

	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

type cli struct {
	a *atomic.Value // address : unixfile
	e *atomic.Value // function error
	i *atomic.Value // function info
	c *atomic.Value // net.Conn
}

func (o *cli) SetTLS(enable bool, config libtls.TLSConfig, serverName string) error {
	return nil
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

func (o *cli) IsConnected() bool {
	if o == nil {
		return false
	} else if i := o.c.Load(); i == nil {
		return false
	} else if c, k := i.(net.Conn); !k {
		return false
	} else if _, e := c.Write(nil); e != nil {
		return false
	} else {
		return true
	}
}

func (o *cli) Connect(ctx context.Context) error {
	if o == nil {
		return ErrInstance
	}

	var (
		err error
		con net.Conn
	)

	o.fctInfo(&net.UnixAddr{}, &net.UnixAddr{}, libsck.ConnectionDial)
	if con, err = o.dial(ctx); err != nil {
		o.fctError(err)
		return err
	}

	o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionNew)
	o.c.Store(con)

	return nil
}

func (o *cli) Read(p []byte) (n int, err error) {
	if o == nil {
		return 0, ErrInstance
	} else if i := o.c.Load(); i == nil {
		return 0, ErrConnection
	} else if c, k := i.(net.Conn); !k {
		return 0, ErrConnection
	} else {
		o.fctInfo(c.LocalAddr(), c.RemoteAddr(), libsck.ConnectionRead)
		return c.Read(p)
	}
}

func (o *cli) Write(p []byte) (n int, err error) {
	if o == nil {
		return 0, ErrInstance
	} else if i := o.c.Load(); i == nil {
		return 0, ErrConnection
	} else if c, k := i.(net.Conn); !k {
		return 0, ErrConnection
	} else {
		o.fctInfo(c.LocalAddr(), c.RemoteAddr(), libsck.ConnectionWrite)
		return c.Write(p)
	}
}

func (o *cli) Close() error {
	if o == nil {
		return ErrInstance
	} else if i := o.c.Load(); i == nil {
		return ErrConnection
	} else if c, k := i.(net.Conn); !k {
		return ErrConnection
	} else {
		o.fctInfo(c.LocalAddr(), c.RemoteAddr(), libsck.ConnectionClose)
		e := c.Close()
		o.c.Store(c)
		return e
	}
}

func (o *cli) Once(ctx context.Context, request io.Reader, fct libsck.Response) error {
	if o == nil {
		return ErrInstance
	}

	defer func() {
		o.fctError(o.Close())
	}()

	var (
		err error
		nbr int64
	)

	if err = o.Connect(ctx); err != nil {
		o.fctError(err)
		return err
	}

	for {
		nbr, err = io.Copy(o, request)

		if err != nil {
			if !errors.Is(err, io.EOF) {
				o.fctError(err)
				return err
			} else {
				break
			}
		} else if nbr < 1 {
			break
		}
	}

	if fct != nil {
		fct(o)
	}

	return nil
}
