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
	"fmt"
	"sync/atomic"
	"time"

	libsrv "github.com/nabbar/golib/server"
)

const (
	pollStop = 500 * time.Millisecond
)

var ErrInvalid = errors.New("invalid instance")

type chData struct {
	cl bool
	ch chan struct{}
}

type run struct {
	e *atomic.Value // slice []error
	f libsrv.Action
	s libsrv.Action
	t *atomic.Value
	c *atomic.Value // chan struct{}
}

func (o *run) getFctStart() libsrv.Action {
	if o.f == nil {
		return func(ctx context.Context) error {
			return fmt.Errorf("invalid start function")
		}
	}
	return o.f
}

func (o *run) getFctStop() libsrv.Action {
	if o.s == nil {
		return func(ctx context.Context) error {
			return fmt.Errorf("invalid stop function")
		}
	}
	return o.s
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

	o.t.Store(time.Time{})

	defer func() {
		libsrv.RecoveryCaller("golib/server/startstop", recover())
		t.Stop()
	}()

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
	if o == nil {
		return nil
	} else {
		return o.getFctStop()(ctx)
	}
}

func (o *run) Start(ctx context.Context) error {
	if e := o.checkMeStart(ctx); e != nil {
		return e
	}

	var can context.CancelFunc
	ctx, can = context.WithCancel(ctx)
	o.t.Store(time.Now())

	go func(x context.Context, n context.CancelFunc) {
		defer n()

		o.chanInit()
		defer func() {
			libsrv.RecoveryCaller("golib/server/startstop", recover())
			_ = o.Stop(ctx)
		}()

		if e := o.getFctStart()(x); e != nil {
			o.errorsAdd(e)
			return
		}
	}(ctx, can)

	go func(x context.Context, n context.CancelFunc) {
		defer n()
		defer func() {
			_ = o.Stop(ctx)
		}()

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
	} else if o.f == nil || o.s == nil {
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

func (o *run) Uptime() time.Duration {
	if i := o.t.Load(); i == nil {
		return 0
	} else if t, k := i.(time.Time); !k {
		return 0
	} else {
		return time.Since(t)
	}
}
