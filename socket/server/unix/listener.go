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
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
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

func (o *srv) buffRead() *bytes.Buffer {
	return bytes.NewBuffer(make([]byte, o.buffSize()))
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
	)

	if f, e = o.getSocketFile(); e != nil {
		return e
	}

	var fctClose = func() {
		if l != nil {
			o.fctError(l.Close())
		}

		if _, e = os.Stat(f); e == nil {
			o.fctError(os.Remove(f))
		}
	}

	if l, e = o.getListen(f); e != nil {
		return e
	}

	defer fctClose()

	// Accept new connection or stop if context or shutdown trigger
	for {
		select {
		case <-ctx.Done():
			return ErrContextClosed
		case <-o.Done():
			return nil
		default:
			// Accept an incoming connection.
			if l == nil {
				return ErrServerClosed
			} else if co, ce := l.Accept(); ce != nil {
				o.fctError(ce)
			} else {
				o.fctInfo(co.LocalAddr(), co.RemoteAddr(), libsck.ConnectionNew)
				go o.Conn(co)
			}
		}
	}
}

func (o *srv) Conn(con net.Conn) {
	defer func() {
		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
		_ = con.Close()
	}()

	o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionNew)

	var (
		err error
		rdr = bufio.NewReaderSize(con, o.buffSize())
		buf []byte
		hdl libsck.Handler
	)

	if hdl = o.handler(); hdl == nil {
		return
	}

	for {
		buf, err = rdr.ReadBytes('\n')

		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionRead)
		if err != nil {
			if err != io.EOF {
				o.fctError(err)
			}
			break
		}

		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionHandler)
		hdl(bytes.NewBuffer(buf), con)
	}
}
