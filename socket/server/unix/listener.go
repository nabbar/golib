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
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

func (o *srv) getSocketFile() (string, error) {
	f := o.sf.Load()
	if f != nil {
		return o.checkFile(f.(string))
	}

	return "", os.ErrNotExist
}

func (o *srv) getSocketPerm() os.FileMode {
	p := o.sp.Load()
	if p > 0 {
		return os.FileMode(p)
	}

	return os.FileMode(0770)
}

func (o *srv) getSocketGroup() int {
	p := o.sg.Load()
	if p >= 0 {
		return int(p)
	}

	gid := syscall.Getgid()
	if gid <= maxGID {
		return gid
	}

	return 0
}

func (o *srv) checkFile(unixFile string) (string, error) {
	if len(unixFile) < 1 {
		return unixFile, fmt.Errorf("missing socket file path")
	} else {
		unixFile = filepath.Join(filepath.Dir(unixFile), filepath.Base(unixFile))
	}

	if _, e := os.Stat(unixFile); e != nil && !errors.Is(e, os.ErrNotExist) {
		return unixFile, e
	} else if e != nil {
		return unixFile, nil
	} else if e = os.Remove(unixFile); e != nil {
		return unixFile, e
	}

	return unixFile, nil
}

func (o *srv) getListen(uxf string) (net.Listener, error) {
	var (
		err error
		prm = o.getSocketPerm()
		grp = o.getSocketGroup()
		old int
		inf fs.FileInfo
		lis net.Listener
	)

	old = syscall.Umask(int(prm))
	defer func() {
		syscall.Umask(old)
	}()

	if lis, err = net.Listen(libptc.NetworkUnix.Code(), uxf); err != nil {
		return nil, err
	} else if inf, err = os.Stat(uxf); err != nil {
		_ = lis.Close()
		return nil, err
	} else if inf.Mode() != prm {
		if err = os.Chmod(uxf, prm); err != nil {
			_ = lis.Close()
			return nil, err
		}
	}

	if stt, ok := inf.Sys().(*syscall.Stat_t); ok {
		if int(stt.Gid) != grp {
			if err = os.Chown(uxf, syscall.Getuid(), grp); err != nil {
				_ = lis.Close()
				return nil, err
			}
		}
	}

	o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkUnix.String(), uxf)
	return lis, nil
}

func (o *srv) Listen(ctx context.Context) error {
	var (
		e error
		f string
		l net.Listener
		s = new(atomic.Bool)
	)

	if f, e = o.getSocketFile(); e != nil {
		o.fctError(e)
		return e
	} else if o.hdl == nil {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidHandler
	} else if l, e = o.getListen(f); e != nil {
		o.fctError(e)
		return e
	}

	s.Store(false)

	defer func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUnix.String(), f)

		if l != nil {
			_ = l.Close()
		}

		if _, e = os.Stat(f); e == nil {
			o.fctError(os.Remove(f))
		}

		o.run.Store(false)
	}()

	o.rst.Store(make(chan struct{}))
	o.stp.Store(make(chan struct{}))
	o.run.Store(true)
	o.gon.Store(false)

	go func() {
		defer func() {
			s.Store(true)

			if l != nil {
				o.fctError(l.Close())
			}

			go func() {
				_ = o.Shutdown(context.Background())
			}()
		}()

		select {
		case <-ctx.Done():
			return
		case <-o.Done():
			return
		}
	}()

	// Accept new connection or stop if context or shutdown trigger
	for l != nil && !s.Load() {
		if co, ce := l.Accept(); ce != nil && !s.Load() {
			o.fctError(ce)
		} else if co != nil {
			o.fctInfo(co.LocalAddr(), co.RemoteAddr(), libsck.ConnectionNew)
			go o.Conn(ctx, co)
		}
	}

	return nil
}

func (o *srv) Conn(ctx context.Context, con net.Conn) {
	var (
		cnl context.CancelFunc
		cor libsck.Reader
		cow libsck.Writer
	)

	o.nc.Add(1) // inc nb connection

	if o.upd != nil {
		o.upd(con)
	}

	ctx, cnl = context.WithCancel(ctx)
	cor, cow = o.getReadWriter(ctx, cnl, con)

	defer func() {
		// cancel context for connection
		cnl()

		// dec nb connection
		o.nc.Add(-1)

		// close connection writer
		_ = cow.Close()

		// delay stopping for 5 seconds to avoid blocking next connection
		if o.IsGone() {
			// if connection is closed
			time.Sleep(5 * time.Second)
		} else {
			// if connection is not closed in 5 seconds
			time.Sleep(500 * time.Millisecond)
		}

		// close connection reader
		_ = cor.Close()

		// send info about connection closing
		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)

		// close connection
		_ = con.Close()
	}()

	// get handler or exit if nil
	if o.hdl == nil {
		return
	} else {
		go o.hdl(cor, cow)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-o.Gone():
			return
		}
	}
}

func (o *srv) getReadWriter(ctx context.Context, cnl context.CancelFunc, con net.Conn) (libsck.Reader, libsck.Writer) {
	var (
		rc = new(atomic.Bool)
		rw = new(atomic.Bool)
	)

	rdrClose := func() error {
		defer func() {
			if rw.Load() {
				cnl()
			}
		}()

		if cr, ok := con.(*net.UnixConn); ok {
			rc.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionCloseRead)
			return libsck.ErrorFilter(cr.CloseRead())
		} else {
			rc.Store(true)
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
			return libsck.ErrorFilter(con.Close())
		}
	}

	wrtClose := func() error {
		defer func() {
			if rc.Load() {
				cnl()
			}
		}()

		if cr, ok := con.(*net.UnixConn); ok {
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionCloseWrite)
			return libsck.ErrorFilter(cr.CloseRead())
		} else {
			rc.Store(true)
			rw.Store(true)
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
			return libsck.ErrorFilter(con.Close())
		}
	}

	rdr := libsck.NewReader(
		func(p []byte) (n int, err error) {
			if ctx.Err() != nil {
				_ = rdrClose()
				return 0, ctx.Err()
			}
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionRead)
			return con.Read(p)
		},
		rdrClose,
		func() bool {
			if ctx.Err() != nil {
				_ = rdrClose()
				return false
			}

			_, e := con.Write(nil)

			if e != nil {
				_ = rdrClose()
				return false
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
				_ = wrtClose()
				return 0, ctx.Err()
			}
			o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionWrite)
			return con.Write(p)
		},
		wrtClose,
		func() bool {
			if ctx.Err() != nil {
				_ = wrtClose()
				return false
			}

			_, e := con.Write(nil)

			if e != nil {
				_ = wrtClose()
				return false
			}

			return true
		},
		func() <-chan struct{} {
			return ctx.Done()
		},
	)

	return rdr, wrt
}
