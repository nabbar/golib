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

package socket

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
)

func (o *srv) timeoutRead() time.Time {
	v := o.tr.Load()
	if v != nil {
		return time.Now().Add(v.(time.Duration))
	}

	return time.Time{}
}

func (o *srv) timeoutWrite() time.Time {
	v := o.tw.Load()
	if v != nil {
		return time.Now().Add(v.(time.Duration))
	}

	return time.Time{}
}

func (o *srv) buffRead() *bytes.Buffer {
	v := o.sr.Load()
	if v > 0 {
		return bytes.NewBuffer(make([]byte, 0, int(v)))
	}

	return bytes.NewBuffer(make([]byte, 0, 32*1024))
}

func (o *srv) buffWrite() *bytes.Buffer {
	v := o.sw.Load()
	if v > 0 {
		return bytes.NewBuffer(make([]byte, 0, int(v)))
	}

	return bytes.NewBuffer(make([]byte, 0, 32*1024))
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

func (o *srv) Listen(ctx context.Context, unixFile string, perm os.FileMode) {
	var (
		e error
		i fs.FileInfo
		l net.Listener
		p = syscall.Umask(int(perm))
	)

	if unixFile, e = o.checkFile(unixFile); e != nil {
		o.fctError(e)
		return
	} else if l, e = net.Listen("unix", unixFile); e != nil {
		o.fctError(e)
		return
	} else if i, e = os.Stat(unixFile); e != nil {
		o.fctError(e)
		return
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
			o.fctError(e)
		}
	}

	// Accept new connection or stop if context or shutdown trigger
	for {
		select {
		case <-ctx.Done():
			return
		case <-o.Done():
			return
		default:
			// Accept an incoming connection.
			if l == nil {
				return
			} else if co, ce := l.Accept(); ce != nil {
				o.fctError(ce)
			} else {
				go o.Conn(co)
			}
		}
	}
}

func (o *srv) Conn(conn net.Conn) {
	defer conn.Close()

	var (
		tr = o.timeoutRead()
		tw = o.timeoutWrite()
		br = o.buffRead()
		bw = o.buffWrite()
	)

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

	if _, e := io.Copy(br, conn); e != nil {
		o.fctError(e)
		return
	} else if e = conn.(*net.UnixConn).CloseRead(); e != nil {
		o.fctError(e)
		return
	}

	if h := o.handler(); h != nil {
		h(br, bw)
	}

	if _, e := io.Copy(conn, bw); e != nil {
		o.fctError(e)
		return
	} else if e = conn.(*net.UnixConn).CloseWrite(); e != nil {
		o.fctError(e)
		return
	}
}
