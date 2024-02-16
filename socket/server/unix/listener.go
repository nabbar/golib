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
		s = new(atomic.Bool)
	)

	if f, e = o.getSocketFile(); e != nil {
		return e
	} else if hdl := o.handler(); hdl == nil {
		return ErrInvalidHandler
	}

	if l, e = o.getListen(f); e != nil {
		return e
	}

	var fctClose = func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUnixGram.String(), f)

		if l != nil {
			_ = l.Close()
		}

		if _, e = os.Stat(f); e == nil {
			o.fctError(os.Remove(f))
		}

		o.r.Store(false)
	}

	defer fctClose()
	s.Store(false)

	go func() {
		<-ctx.Done()
		go func() {
			_ = o.Shutdown()
		}()
		return
	}()

	go func() {
		<-o.Done()

		e = nil
		s.Store(true)

		if l != nil {
			o.fctError(l.Close())
		}

		return
	}()

	o.r.Store(true)
	// Accept new connection or stop if context or shutdown trigger
	for {
		// Accept an incoming connection.
		if l == nil {
			return ErrServerClosed
		} else if s.Load() {
			return e
		}

		if co, ce := l.Accept(); ce != nil && !s.Load() {
			o.fctError(ce)
		} else if co != nil {
			o.fctInfo(co.LocalAddr(), co.RemoteAddr(), libsck.ConnectionNew)
			go o.Conn(co)
		}
	}
}

func (o *srv) Conn(con net.Conn) {
	defer func() {
		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionClose)
		_ = con.Close()
	}()

	var (
		err error
		rdr = bufio.NewReaderSize(con, o.buffSize())
		msg []byte
		hdl libsck.Handler
	)

	if hdl = o.handler(); hdl == nil {
		return
	}

	for {
		msg, err = rdr.ReadBytes('\n')

		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionRead)
		if err != nil {
			if err != io.EOF {
				o.fctError(err)
			}
			if len(msg) < 1 {
				break
			}
		}

		var buf = bytes.NewBuffer(msg)

		if !bytes.HasSuffix(msg, []byte{libsck.EOL}) {
			buf.Write([]byte{libsck.EOL})
		}

		o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionHandler)
		hdl(buf, con)
	}
}
