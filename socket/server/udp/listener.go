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

func (o *srv) getListen(addr string) (*net.UDPConn, error) {
	var (
		adr *net.UDPAddr
		lis *net.UDPConn
		err error
	)

	if adr, err = net.ResolveUDPAddr(libptc.NetworkUDP.Code(), addr); err != nil {
		return nil, err
	} else if lis, err = net.ListenUDP(libptc.NetworkUDP.Code(), adr); err != nil {
		return lis, err
	} else {
		o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkUDP.String(), addr)
	}

	return lis, nil
}

func (o *srv) Listen(ctx context.Context) error {
	var (
		err error
		nbr int
		loc *net.UDPAddr
		stp = new(atomic.Bool)
		rem net.Addr
		con *net.UDPConn
		adr = o.getAddress()
		hdl libsck.Handler
	)

	if len(adr) == 0 {
		return ErrInvalidAddress
	} else if hdl = o.handler(); hdl == nil {
		return ErrInvalidHandler
	} else if loc, err = net.ResolveUDPAddr(libptc.NetworkUDP.Code(), adr); err != nil {
		return err
	}

	if con, err = net.ListenUDP(libptc.NetworkUDP.Code(), loc); err != nil {
		o.fctError(err)
		return err
	}

	var fctClose = func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUDP.String(), adr)

		if con != nil {
			_ = con.Close()
		}

		o.r.Store(false)
	}

	defer fctClose()
	stp.Store(false)

	go func() {
		<-ctx.Done()
		go func() {
			_ = o.Shutdown()
		}()
		return
	}()

	go func() {
		<-o.Done()

		err = nil
		stp.Store(true)

		if con != nil {
			o.fctError(con.Close())
		}

		return
	}()

	o.r.Store(true)

	o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkUDP.String(), adr)
	// Accept new connection or stop if context or shutdown trigger

	var (
		siz = o.buffSize()
		buf []byte
		rer error
	)

	for {
		// Accept an incoming connection.
		if con == nil {
			return ErrServerClosed
		} else if stp.Load() {
			return err
		}

		buf = make([]byte, siz)
		nbr, rem, rer = con.ReadFrom(buf)

		if rem == nil {
			rem = &net.UDPAddr{}
		}

		o.fctInfo(loc, rem, libsck.ConnectionRead)

		if rer != nil {
			if !stp.Load() {
				o.fctError(rer)
			}
		}

		if nbr > 0 {
			if !bytes.HasSuffix(buf, []byte{libsck.EOL}) {
				buf = append(buf, libsck.EOL)
				nbr++
			}

			o.fctInfo(loc, rem, libsck.ConnectionHandler)
			hdl(bytes.NewBuffer(buf[:nbr]), io.Discard)
		}
	}
}
