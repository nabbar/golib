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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"

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
		err error
		nbr int
		uxf string
		stp = new(atomic.Bool)
		loc *net.UnixAddr
		rem net.Addr
		con *net.UnixConn
		hdl libsck.Handler
	)

	if uxf, err = o.getSocketFile(); err != nil {
		return err
	} else if hdl = o.handler(); hdl == nil {
		return ErrInvalidHandler
	} else if loc, err = net.ResolveUnixAddr(libptc.NetworkUnixGram.Code(), uxf); err != nil {
		return err
	} else if con, err = o.getListen(uxf, loc); err != nil {
		return err
	}

	var fctClose = func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUnixGram.String(), uxf)

		if con != nil {
			_ = con.Close()
		}

		if _, err = os.Stat(uxf); err == nil {
			o.fctError(os.Remove(uxf))
		}
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

	var (
		siz = o.buffSize()
		buf []byte
		rer error
	)

	// Accept new connection or stop if context or shutdown trigger
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
			rem = &net.UnixAddr{}
		}

		o.fctInfo(loc, rem, libsck.ConnectionRead)

		if nbr > 0 {
			if !bytes.HasSuffix(buf, []byte{libsck.EOL}) {
				buf = append(buf, libsck.EOL)
				nbr++
			}

			o.fctInfo(loc, rem, libsck.ConnectionHandler)
			hdl(bytes.NewBuffer(buf[:nbr]), io.Discard)
		}

		if rer != nil {
			if !stp.Load() {
				o.fctError(rer)
			}
		}
	}
}
