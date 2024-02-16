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
	"os"
	"sync/atomic"
	"time"
)

const (
	pollStop        = 500 * time.Millisecond
	defaultDuration = 30 * time.Second
)

var ErrInvalid = errors.New("invalid instance")

type chData struct {
	cl bool
	ch chan struct{}
}

type run struct {
	e *atomic.Value // slice []error
	f func(ctx context.Context, tck *time.Ticker) error
	d time.Duration
	t *atomic.Value
	c *atomic.Value // chan struct{}
}

func (o *run) getDuration() time.Duration {
	// still check on function checkMe
	return o.d
}

func (o *run) getFunction() func(ctx context.Context, tck *time.Ticker) error {
	// still check on function checkMe
	return o.f
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
	}

	o.t.Store(time.Time{})

	if !o.IsRunning() {
		return nil
	}

	var t = time.NewTicker(pollStop)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if o.IsRunning() {
				o.chanSend()
			} else {
				return nil
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

	go func(con context.Context) {
		var (
			tck  = time.NewTicker(o.getDuration())
			x, n = context.WithCancel(con)
		)

		defer func() {
			if rec := recover(); rec != nil {
				_, _ = fmt.Fprintf(os.Stderr, "recovering panic thread on Start function in gollib/server/ticker/model.\n%v\n", rec)
			}
			if n != nil {
				n()
			}
			if tck != nil {
				tck.Stop()
			}
			o.chanClose()
		}()

		o.t.Store(time.Now())

		o.chanInit()
		o.errorsClean()

		for {
			select {
			case <-tck.C:
				f := func(ctx context.Context, tck *time.Ticker) error {
					defer func() {
						if rec := recover(); rec != nil {
							_, _ = fmt.Fprintf(os.Stderr, "recovering panic while calling function.\n%v\n", rec)
						}
					}()
					return o.getFunction()(ctx, tck)
				}
				if e := f(x, tck); e != nil {
					o.errorsAdd(e)
				}
			case <-con.Done():
				return
			case <-o.chanDone():
				return
			}
		}
	}(ctx)

	return nil
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
