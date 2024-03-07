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
	f libsrv.FuncAction
	s libsrv.FuncAction
	t *atomic.Value
	r *atomic.Bool  // Want Stop
	n *atomic.Value // context.CancelFunc
}

func (o *run) getFctStart() libsrv.FuncAction {
	if o.f == nil {
		return func(ctx context.Context) error {
			return fmt.Errorf("invalid start function")
		}
	}
	return o.f
}

func (o *run) getFctStop() libsrv.FuncAction {
	if o.s == nil {
		return func(ctx context.Context) error {
			return fmt.Errorf("invalid stop function")
		}
	}
	return o.s
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

func (o *run) IsRunning() bool {
	if e := o.checkMe(); e != nil {
		return false
	}

	return o.r.Load()
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

	o.errorsClean()

	if i := o.n.Load(); i != nil {
		if f, k := i.(context.CancelFunc); k {
			f()
		}
	}

	defer func() {
		libsrv.RecoveryCaller("golib/server/startstop", recover())
		t.Stop()
	}()

	for {
		select {
		case <-t.C:
			if o.IsRunning() {
				e = o.getFctStop()(ctx)
			} else {
				return e
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (o *run) Start(ctx context.Context) error {
	if e := o.checkMeStart(ctx); e != nil {
		return e
	}

	cx, ca := context.WithCancel(ctx)

	if i := o.n.Swap(ca); i != nil {
		if f, k := i.(context.CancelFunc); k {
			f()
		}
	}

	o.errorsClean()

	fStart := func(c context.Context) error {
		defer libsrv.RecoveryCaller("golib/server/startstop", recover())
		return o.getFctStart()(c)
	}

	fStop := func(c context.Context) error {
		defer libsrv.RecoveryCaller("golib/server/startstop", recover())
		return o.getFctStop()(c)
	}

	go func(x context.Context, n context.CancelFunc, start, stop libsrv.FuncAction) {
		defer func() {
			libsrv.RecoveryCaller("golib/server/startstop", recover())
			_ = stop(ctx)

			n()

			o.t.Store(time.Time{})
			o.r.Store(false)
		}()

		o.t.Store(time.Now())

		for !o.r.Load() && ctx.Err() == nil {
			o.r.Store(true)

			if e := start(x); e != nil {
				o.errorsAdd(e)
				return
			}
		}
	}(cx, ca, fStart, fStop)

	return nil
}
