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
	"context"
	"crypto/tls"
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
	ssl *atomic.Value // tls config
	hdl *atomic.Value // handler
	msg *atomic.Value // chan []byte
	stp *atomic.Value // chan struct{}
	rst *atomic.Value // chan struct{}
	run *atomic.Bool  // is Running
	gon *atomic.Bool  // is Running

	fe *atomic.Value // function error
	fi *atomic.Value // function info
	fs *atomic.Value // function info server

	sr *atomic.Int32 // read buffer size
	ad *atomic.Value // Server address url

	nc *atomic.Int64 // Counter Connection
}

func (o *srv) OpenConnections() int64 {
	return o.nc.Load()
}

func (o *srv) IsRunning() bool {
	return o.run.Load()
}

func (o *srv) IsGone() bool {
	return o.gon.Load()
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
	if o == nil {
		return closedChanStruct
	}
	if o.IsGone() {
		return closedChanStruct
	} else if i := o.rst.Load(); i != nil {
		if g, k := i.(chan struct{}); k {
			return g
		}
	}

	return closedChanStruct
}

func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

func (o *srv) StopGone(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}

	o.gon.Store(true)

	if i := o.rst.Load(); i != nil {
		if c, k := i.(chan struct{}); k && c != closedChanStruct {
			close(c)
		}
	}
	o.rst.Store(closedChanStruct)

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
			return ErrGoneTimeout
		case <-tck.C:
			if o.OpenConnections() > 0 {
				continue
			}
			return nil
		}
	}

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

	e := o.StopGone(ctx)
	if err := o.StopListen(ctx); err != nil {
		return err
	} else {
		return e
	}
}

func (o *srv) SetTLS(enable bool, config libtls.TLSConfig) error {
	if !enable {
		// #nosec
		o.ssl.Store(&tls.Config{})
		return nil
	}

	if config == nil {
		return fmt.Errorf("invalid tls config")
	} else if l := config.GetCertificatePair(); len(l) < 1 {
		return fmt.Errorf("invalid tls config, missing certificates pair")
	} else if t := config.TlsConfig(""); t == nil {
		return fmt.Errorf("invalid tls config")
	} else {
		o.ssl.Store(t)
		return nil
	}
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
	} else if _, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), address); err != nil {
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

func (o *srv) getTLS() *tls.Config {
	i := o.ssl.Load()

	if i == nil {
		return nil
	} else if t, k := i.(*tls.Config); !k {
		return nil
	} else if len(t.Certificates) < 1 {
		return nil
	} else {
		return t
	}
}
