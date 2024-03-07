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

package ticker

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	libsrv "github.com/nabbar/golib/server"
)

const (
	pollStop        = 500 * time.Millisecond
	defaultDuration = 30 * time.Second
)

var ErrInvalid = errors.New("invalid instance")

type run struct {
	e *atomic.Value // slice []error
	f libsrv.FuncTicker
	d time.Duration
	t *atomic.Value
	r *atomic.Bool
	n *atomic.Value
}

func (o *run) getDuration() time.Duration {
	// still check on function checkMe
	return o.d
}

func (o *run) getFunction() libsrv.FuncTicker {
	// still check on function checkMe
	if o.f == nil {
		return func(ctx context.Context, tck *time.Ticker) error {
			return fmt.Errorf("invalid function ticker")
		}
	}

	return o.f
}

func (o *run) checkMe() error {
	if o == nil {
		return ErrInvalid
	} else if o.f == nil || o.d == 0 {
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

	var t = time.NewTicker(pollStop)
	defer t.Stop()

	o.errorsClean()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			if o.IsRunning() {
				if i := o.n.Load(); i != nil {
					if f, k := i.(context.CancelFunc); k {
						f()
					} else {
						o.r.Store(false)
					}
				} else {
					o.r.Store(false)
				}
			} else {
				return nil
			}
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

	fct := func(ctx context.Context, tck *time.Ticker) error {
		defer libsrv.RecoveryCaller("golib/server/ticker", recover())
		return o.getFunction()(ctx, tck)
	}

	go func(x context.Context, n context.CancelFunc, f libsrv.FuncTicker) {
		var tck = time.NewTicker(o.getDuration())

		defer func() {
			libsrv.RecoveryCaller("golib/server/ticker", recover())
			tck.Stop()
			n()

			o.t.Store(time.Time{})
			o.r.Store(false)
		}()

		o.r.Store(true)
		o.t.Store(time.Now())

		for {
			select {
			case <-x.Done():
				return
			case <-tck.C:
				o.r.Store(true)
				if e := f(x, tck); e != nil {
					o.errorsAdd(e)
				}
			}
		}
	}(cx, ca, fct)

	return nil
}
