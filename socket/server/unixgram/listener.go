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
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"

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

func (o *srv) getListen(uxf string, adr *net.UnixAddr) (*net.UnixConn, error) {
	var (
		err error
		prm = o.getSocketPerm()
		grp = o.getSocketGroup()
		old int
		inf fs.FileInfo
		lis *net.UnixConn
	)

	old = syscall.Umask(int(prm))
	defer func() {
		syscall.Umask(old)
	}()

	if lis, err = net.ListenUnixgram(libptc.NetworkUnixGram.Code(), adr); err != nil {
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

	o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkUnixGram.String(), uxf)
	return lis, nil
}

func (o *srv) Listen(ctx context.Context) error {
	var (
		e error
		u string
		s = new(atomic.Bool)

		loc *net.UnixAddr
		con *net.UnixConn
		cnl context.CancelFunc
		cor libsck.Reader
		cow libsck.Writer
	)

	s.Store(false)

	if u, e = o.getSocketFile(); e != nil {
		o.fctError(e)
		return e
	} else if o.hdl == nil {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidHandler
	} else if loc, e = net.ResolveUnixAddr(libptc.NetworkUnixGram.Code(), u); e != nil {
		o.fctError(e)
		return e
	} else if con, e = o.getListen(u, loc); e != nil {
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
		o.fctInfo(loc, &net.UnixAddr{}, libsck.ConnectionClose)
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUnixGram.String(), u)

		// close connection
		_ = con.Close()

		if _, e = os.Stat(u); e == nil {
			o.fctError(os.Remove(u))
		}

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

func (o *srv) getReadWriter(ctx context.Context, con *net.UnixConn, loc net.Addr) (libsck.Reader, libsck.Writer) {
	var (
		re = &net.UDPAddr{}
		ra = new(atomic.Value)
		fg = func() net.Addr {
			if i := ra.Load(); i != nil {
				if v, k := i.(net.Addr); k {
					return v
				}
			}
			return &net.UnixAddr{}
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
