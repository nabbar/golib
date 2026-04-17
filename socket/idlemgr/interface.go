/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package idlemgr

import (
	"context"
	"errors"
	"io"

	durbig "github.com/nabbar/golib/duration/big"
	libsrv "github.com/nabbar/golib/runner"
	runtck "github.com/nabbar/golib/runner/ticker"
)

var (
	ErrInvalidInstance = errors.New("invalid instance")
	ErrInvalidClient   = errors.New("invalid client")
)

type Client interface {
	io.Closer

	Ref() string
	Inc()
	Get() uint32
}

type Manager interface {
	libsrv.Runner
	io.Closer

	Register(Client) error
	Unregister(Client) error
}

func New(ctx context.Context, idle, tick durbig.Duration) (Manager, error) {
	t, e := tick.Time()
	if e != nil {
		return nil, e
	}

	m := &mgr{
		x: ctx,
		i: idle.Uint32(),
	}

	for i := 0; i < numShards; i++ {
		m.s[i].c = make(map[string]Client)
	}

	m.r = runtck.New(t, m.run)
	return m, nil
}
