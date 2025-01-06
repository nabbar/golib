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
	"context"
	"net"
	"sync/atomic"

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

func (o *srv) getListen(addr string) (*net.UDPAddr, *net.UDPConn, error) {
	var (
		adr *net.UDPAddr
		lis *net.UDPConn
		err error
	)

	if adr, err = net.ResolveUDPAddr(libptc.NetworkUDP.Code(), addr); err != nil {
		return nil, nil, err
	} else if lis, err = net.ListenUDP(libptc.NetworkUDP.Code(), adr); err != nil {
		if lis != nil {
			_ = lis.Close()
		}
		return nil, nil, err
	} else {
		o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkUDP.String(), addr)
	}

	return adr, lis, nil
}

func (o *srv) Listen(ctx context.Context) error {
	var (
		e error
		s = new(atomic.Bool)
		a = o.getAddress()

		loc *net.UDPAddr
		con *net.UDPConn
		cnl context.CancelFunc
		cor libsck.Reader
		cow libsck.Writer
	)

	s.Store(false)

	if len(a) == 0 {
		o.fctError(ErrInvalidInstance)
		return ErrInvalidAddress
	} else if o.hdl == nil {
		o.fctError(ErrInvalidInstance)
		return ErrInvalidHandler
	} else if loc, con, e = o.getListen(a); e != nil {
		o.fctError(e)
		return e
	}

	if o.upd != nil {
		o.upd(con)
	}

	ctx, cnl = context.WithCancel(ctx)
	cor, cow = o.getReadWriter(ctx, con, loc)

	o.stp.Store(make(chan struct{}))
	o.run.Store(true)

	defer func() {
		// cancel context for connection
		cnl()

		// send info about connection closing
		o.fctInfo(loc, &net.UDPAddr{}, libsck.ConnectionClose)
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUDP.String(), a)

		// close connection
		_ = con.Close()

		o.run.Store(false)
	}()

	// get handler or exit if nil
	go o.hdl(cor, cow)

	for {
		select {
		case <-ctx.Done():
			return ErrContextClosed
		case <-o.Done():
			return nil
		}
	}
}

func (o *srv) getReadWriter(ctx context.Context, con *net.UDPConn, loc net.Addr) (libsck.Reader, libsck.Writer) {
	var (
		re = &net.UDPAddr{}
		ra = new(atomic.Value)
		fg = func() net.Addr {
			if i := ra.Load(); i != nil {
				if v, k := i.(net.Addr); k {
					return v
				}
			}
			return &net.UDPAddr{}
		}
	)
	ra.Store(re)

	fctClose := func() error {
		o.fctInfo(loc, fg(), libsck.ConnectionClose)
		return libsck.ErrorFilter(con.Close())
	}

	rdr := libsck.NewReader(
		func(p []byte) (n int, err error) {
			if ctx.Err() != nil {
				_ = fctClose()
				return 0, ctx.Err()
			}

			var a net.Addr
			n, a, err = con.ReadFrom(p)

			if a != nil {
				ra.Store(a)
			} else {
				ra.Store(re)
			}

			o.fctInfo(loc, fg(), libsck.ConnectionRead)
			return n, err
		},
		fctClose,
		func() bool {
			if ctx.Err() != nil {
				_ = fctClose()
				return false
			}
			_, e := con.Read(nil)

			if e != nil {
				_ = fctClose()
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
				_ = fctClose()
				return 0, ctx.Err()
			}

			if a := fg(); a != nil && a != re {
				o.fctInfo(loc, a, libsck.ConnectionWrite)
				return con.WriteTo(p, a)
			}

			o.fctInfo(loc, fg(), libsck.ConnectionWrite)
			return con.Write(p)
		},
		fctClose,
		func() bool {
			if ctx.Err() != nil {
				_ = fctClose()
				return false
			}

			_, e := con.Write(nil)

			if e != nil {
				_ = fctClose()
			}

			return true
		},
		func() <-chan struct{} {
			return ctx.Done()
		},
	)

	return rdr, wrt
}
