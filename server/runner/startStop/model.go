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

package startStop

import (
	"context"
	"errors"
	"sync"
	"time"

	libsrv "github.com/nabbar/golib/server"
)

const (
	pollStop = 500 * time.Millisecond
)

var ErrInvalid = errors.New("invalid instance")

type run struct {
	m sync.RWMutex
	e []error
	f func(ctx context.Context) error
	s func(ctx context.Context) error
	c chan struct{}
}

func (o *run) Restart(ctx context.Context) error {
	if e := o.Stop(ctx); e != nil {
		return e
	} else if e = o.Start(ctx); e != nil {
		_ = o.Stop(ctx)
		return e
	}

	return nil
}

func (o *run) Stop(ctx context.Context) error {
	if e := o.checkMe(); e != nil {
		return e
	} else if !o.IsRunning() {
		return nil
	}

	var (
		e error
		t = time.NewTicker(pollStop)
	)

	defer t.Stop()

	for {
		select {
		case <-t.C:
			if o.IsRunning() {
				o.chanSend()
				e = o.callStop(ctx)
			} else {
				return e
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (o *run) callStop(ctx context.Context) error {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.s == nil {
		return nil
	}

	return o.s(ctx)
}

func (o *run) Start(ctx context.Context) error {
	if e := o.checkMeStart(ctx); e != nil {
		return e
	}

	o.m.RLock()
	defer o.m.RUnlock()

	var can context.CancelFunc
	ctx, can = context.WithCancel(ctx)

	go func(x context.Context, n context.CancelFunc, fct libsrv.Action) {
		defer n()

		o.chanInit()
		defer o.chanClose()

		if e := fct(x); e != nil {
			o.errorsAdd(e)
			return
		}
	}(ctx, can, o.f)

	go func(x context.Context, n context.CancelFunc) {
		defer n()
		defer o.chanClose()

		for {
			select {
			case <-o.chanDone():
				return
			case <-x.Done():
				return
			}
		}
	}(ctx, can)

	return nil
}

func (o *run) checkMe() error {
	if o == nil {
		return ErrInvalid
	}

	o.m.RLock()
	defer o.m.RUnlock()

	if o.f == nil || o.s == nil {
		return ErrInvalid
	}

	return nil
}

func (o *run) checkMeStart(ctx context.Context) error {
	if e := o.checkMe(); e != nil {
		return e
	} else if o.IsRunning() {
		if e = o.Stop(ctx); e != nil {
			return e
		}
	}

	return nil
}
