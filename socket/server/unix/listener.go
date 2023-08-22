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
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

func (o *srv) timeoutRead() time.Time {
	v := o.tr.Load()
	if v == nil {
		return time.Time{}
	} else if d, k := v.(time.Duration); !k {
		return time.Time{}
	} else if d > 0 {
		return time.Now().Add(v.(time.Duration))
	}

	return time.Time{}
}

func (o *srv) timeoutWrite() time.Time {
	v := o.tw.Load()
	if v == nil {
		return time.Time{}
	} else if d, k := v.(time.Duration); !k {
		return time.Time{}
	} else if d > 0 {
		return time.Now().Add(v.(time.Duration))
	}

	return time.Time{}
}

func (o *srv) buffRead() *bytes.Buffer {
	v := o.sr.Load()
	if v > 0 {
		return bytes.NewBuffer(make([]byte, 0, int(v)))
	}

	return bytes.NewBuffer(make([]byte, 0, libsck.DefaultBufferSize))
}

func (o *srv) buffWrite() *bytes.Buffer {
	v := o.sw.Load()
	if v > 0 {
		return bytes.NewBuffer(make([]byte, 0, int(v)))
	}

	return bytes.NewBuffer(make([]byte, 0, libsck.DefaultBufferSize))
}

func (o *srv) getSocketFile() (string, error) {
	f := o.fs.Load()
	if f != nil {
		return o.checkFile(f.(string))
	}

	return "", os.ErrNotExist
}

func (o *srv) getSocketPerm() os.FileMode {
	p := o.fp.Load()
	if p > 0 {
		return os.FileMode(p)
	}

	return os.FileMode(0770)
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

func (o *srv) Listen(ctx context.Context) error {
	var (
		e error
		i fs.FileInfo
		l net.Listener

		perm     = o.getSocketPerm()
		unixFile string

		p = syscall.Umask(int(perm))
	)

	if unixFile, e = o.getSocketFile(); e != nil {
		return e
	} else if l, e = net.Listen(libptc.NetworkUnix.Code(), unixFile); e != nil {
		return e
	} else if i, e = os.Stat(unixFile); e != nil {
		return e
	}

	syscall.Umask(p)

	var fctClose = func() {
		if l != nil {
			o.fctError(l.Close())
		}

		if i, e = os.Stat(unixFile); e == nil {
			o.fctError(os.Remove(unixFile))
		}
	}

	defer fctClose()

	if i.Mode() != perm {
		if e = os.Chmod(unixFile, perm); e != nil {
			return e
		}
	}

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
				go o.Conn(co)
			}
		}
	}
}

func (o *srv) Conn(conn net.Conn) {
	defer func() {
		e := conn.Close()
		o.fctError(e)
	}()

	var (
		lc = conn.LocalAddr()
		rm = conn.RemoteAddr()
		tr = o.timeoutRead()
		tw = o.timeoutWrite()
		br = o.buffRead()
		bw = o.buffWrite()
	)

	defer o.fctInfo(lc, rm, libsck.ConnectionClose)

	o.fctInfo(lc, rm, libsck.ConnectionNew)

	if !tr.IsZero() {
		if e := conn.SetReadDeadline(tr); e != nil {
			o.fctError(e)
			return
		}
	}

	if !tw.IsZero() {
		if e := conn.SetReadDeadline(tw); e != nil {
			o.fctError(e)
			return
		}
	}

	o.fctInfo(lc, rm, libsck.ConnectionRead)

	if _, e := io.Copy(br, conn); e != nil {
		o.fctError(e)
		return
	}

	o.fctInfo(lc, rm, libsck.ConnectionCloseRead)

	if e := conn.(*net.UnixConn).CloseRead(); e != nil {
		o.fctError(e)
		return
	}

	if h := o.handler(); h != nil {
		o.fctInfo(lc, rm, libsck.ConnectionHandler)
		h(br, bw)
	}

	if bw.Len() > 0 {
		o.fctInfo(lc, rm, libsck.ConnectionWrite)
		if _, e := io.Copy(conn, bw); e != nil {
			o.fctError(e)
		}
	}

	o.fctInfo(lc, rm, libsck.ConnectionCloseWrite)
	if e := conn.(*net.UnixConn).CloseWrite(); e != nil {
		o.fctError(e)
	}
}
