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
	"fmt"
	"net"
	"sync/atomic"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

var (
	closedChanStruct chan struct{}
)

func init() {
	closedChanStruct = make(chan struct{})
	close(closedChanStruct)
}

type srv struct {
	hdl *atomic.Value // handler
	msg *atomic.Value // chan []byte
	stp *atomic.Value // chan struct{}
	run *atomic.Bool  // is Running

	fe *atomic.Value // function error
	fi *atomic.Value // function info
	fs *atomic.Value // function info server

	ad *atomic.Value // Server address url
}

func (o *srv) OpenConnections() int64 {
	if o.IsRunning() {
		return 1
	}

	return 0
}

func (o *srv) IsRunning() bool {
	return o.run.Load()
}

func (o *srv) IsGone() bool {
	return !o.IsRunning()
}

func (o *srv) Done() <-chan struct{} {
	if o == nil {
		return closedChanStruct
	}

	if i := o.stp.Load(); i != nil {
		if c, k := i.(chan struct{}); k {
			return c
		}
	}

	return closedChanStruct
}

func (o *srv) Gone() <-chan struct{} {
	return closedChanStruct
}

func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

func (o *srv) StopListen(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}

	if i := o.stp.Load(); i != nil {
		if c, k := i.(chan struct{}); k && c != closedChanStruct {
			close(c)
		}
	}
	o.stp.Store(closedChanStruct)

	var (
		tck = time.NewTicker(5 * time.Millisecond)
		cnl context.CancelFunc
	)

	ctx, cnl = context.WithTimeout(ctx, 10*time.Second)

	defer func() {
		tck.Stop()
		cnl()
	}()

	for {
		select {
		case <-ctx.Done():
			return ErrShutdownTimeout
		case <-tck.C:
			if o.IsRunning() {
				continue
			}
			return nil
		}
	}

}

func (o *srv) Shutdown(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}

	var cnl context.CancelFunc
	ctx, cnl = context.WithTimeout(ctx, 25*time.Second)
	defer cnl()
	return o.StopListen(ctx)
}

func (o *srv) SetTLS(enable bool, config libtls.TLSConfig) error {
	return nil
}

func (o *srv) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.fe.Store(f)
}

func (o *srv) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	o.fi.Store(f)
}

func (o *srv) RegisterFuncInfoServer(f libsck.FuncInfoSrv) {
	if o == nil {
		return
	}

	o.fs.Store(f)
}

func (o *srv) RegisterServer(address string) error {
	if len(address) < 1 {
		return ErrInvalidAddress
	} else if _, err := net.ResolveUDPAddr(libptc.NetworkUDP.Code(), address); err != nil {
		return err
	}

	o.ad.Store(address)
	return nil
}

func (o *srv) fctError(e error) {
	if o == nil {
		return
	}

	v := o.fe.Load()
	if v != nil {
		v.(libsck.FuncError)(e)
	}
}

func (o *srv) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	if o == nil {
		return
	}

	v := o.fi.Load()
	if v != nil {
		v.(libsck.FuncInfo)(local, remote, state)
	}
}

func (o *srv) fctInfoSrv(msg string, args ...interface{}) {
	if o == nil {
		return
	}

	v := o.fs.Load()
	if v != nil {
		v.(libsck.FuncInfoSrv)(fmt.Sprintf(msg, args...))
	}
}

func (o *srv) handler() libsck.Handler {
	if o == nil {
		return nil
	}

	v := o.hdl.Load()
	if v != nil {
		return v.(libsck.Handler)
	}

	return nil
}
