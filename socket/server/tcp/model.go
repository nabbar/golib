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
	"crypto/tls"
	"fmt"
	"net"
	"sync/atomic"

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

type data struct {
	data any
}

type srv struct {
	l net.Listener

	t *atomic.Value // tls config
	h *atomic.Value // handler
	c *atomic.Value // chan []byte
	s *atomic.Value // chan struct{}

	fe *atomic.Value // function error
	fi *atomic.Value // function info
	fs *atomic.Value // function info server

	sr *atomic.Int32 // read buffer size
	ad *atomic.Value // Server address url
}

func (o *srv) Done() <-chan struct{} {
	s := o.s.Load()
	if s != nil {
		return s.(chan struct{})
	}

	return closedChanStruct
}

func (o *srv) Shutdown() {
	if o == nil {
		return
	}

	s := o.s.Load()
	if s != nil {
		o.s.Store(nil)
	}
}

func (o *srv) SetTLS(enable bool, config libtls.TLSConfig) error {
	if !enable {
		// #nosec
		o.t.Store(&tls.Config{})
		return nil
	}

	if config == nil {
		return fmt.Errorf("invalid tls config")
	} else if l := config.GetCertificatePair(); len(l) < 1 {
		return fmt.Errorf("invalid tls config, missing certificates pair")
	} else if t := config.TlsConfig(""); t == nil {
		return fmt.Errorf("invalid tls config")
	} else {
		o.t.Store(t)
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

	v := o.h.Load()
	if v != nil {
		return v.(libsck.Handler)
	}

	return nil
}

func (o *srv) getTLS() *tls.Config {
	i := o.t.Load()

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
