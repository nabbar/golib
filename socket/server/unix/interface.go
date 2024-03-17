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
	"os"
	"sync/atomic"

	libsck "github.com/nabbar/golib/socket"
)

const maxGID = 32767

type ServerUnix interface {
	libsck.Server
	RegisterSocket(unixFile string, perm os.FileMode, gid int32) error
}

func New(h libsck.Handler) ServerUnix {
	c := new(atomic.Value)
	c.Store(make(chan []byte))

	s := new(atomic.Value)
	s.Store(make(chan struct{}))

	r := new(atomic.Value)
	r.Store(make(chan struct{}))

	f := new(atomic.Value)
	f.Store(h)

	// socket file
	sf := new(atomic.Value)
	sf.Store("")

	// socket permission
	sp := new(atomic.Int64)
	sp.Store(0)

	// socket group permission
	sg := new(atomic.Int32)
	sg.Store(0)

	return &srv{
		hdl: f,
		msg: c,
		stp: s,
		rst: r,
		run: new(atomic.Bool),
		gon: new(atomic.Bool),
		fe:  new(atomic.Value),
		fi:  new(atomic.Value),
		fs:  new(atomic.Value),
		sf:  sf,
		sp:  sp,
		sg:  sg,
		nc:  new(atomic.Int64),
	}
}
