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

package unixgram

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"

	libtls "github.com/nabbar/golib/certificates"
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
	l net.Listener

	h *atomic.Value // handler
	c *atomic.Value // chan []byte
	s *atomic.Value // chan struct{}
	r *atomic.Bool  // is Running

	fe *atomic.Value // function error
	fi *atomic.Value // function info
	fs *atomic.Value // function info server

	sr *atomic.Int32 // read buffer size
	sf *atomic.Value // file unix socket
	sp *atomic.Int64 // file unix perm
	sg *atomic.Int32 // file unix group perm
}

func (o *srv) IsRunning() bool {
	return o.r.Load()
}

func (o *srv) Done() <-chan struct{} {
	s := o.s.Load()
	if s != nil {
		return s.(chan struct{})
	}

	return closedChanStruct
}

func (o *srv) Close() error {
	return o.Shutdown()
}

func (o *srv) Shutdown() error {
	if o == nil {
		return ErrInvalidInstance
	}

	s := o.s.Load()
	if s != nil {
		if c, k := s.(chan struct{}); k {
			c <- struct{}{}
			o.s.Store(c)
		}
	}

	var (
		tck      = time.NewTicker(100 * time.Millisecond)
		ctx, cnl = context.WithTimeout(context.Background(), 25*time.Second)
	)

	defer func() {
		if s != nil {
			o.s.Store(closedChanStruct)
		}

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

func (o *srv) RegisterSocket(unixFile string, perm os.FileMode, gid int32) error {
	if _, err := net.ResolveUnixAddr(libptc.NetworkUnixGram.Code(), unixFile); err != nil {
		return err
	} else if gid > maxGID {
		return ErrInvalidGroup
	}

	o.sf.Store(unixFile)
	o.sp.Store(int64(perm))
	o.sg.Store(gid)

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

	v := o.h.Load()
	if v != nil {
		return v.(libsck.Handler)
	}

	return nil
}
