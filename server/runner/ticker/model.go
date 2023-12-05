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
	"sync"
	"time"
)

const (
	pollStop = 500 * time.Millisecond
)

var ErrInvalid = errors.New("invalid instance")

type run struct {
	m sync.RWMutex
	e []error
	f func(ctx context.Context, tck *time.Ticker) error
	d time.Duration
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

	o.m.RLock()
	defer o.m.RUnlock()

	go func(con context.Context, dur time.Duration, fct func(ctx context.Context, tck *time.Ticker) error) {
		var (
			tck  = time.NewTicker(dur)
			x, n = context.WithCancel(con)
		)

		defer func() {
			if rec := recover(); rec != nil {
				_, _ = fmt.Fprintf(os.Stderr, "recovering panic thread on gollib/server/ticker/model.\n%v\n", rec)
			}
			if n != nil {
				n()
			}
			if tck != nil {
				tck.Stop()
			}
			o.chanClose()
		}()

		o.chanInit()
		o.errorsClean()

		for {
			select {
			case <-tck.C:
				f := func(ctx context.Context, tck *time.Ticker) error {
					defer func() {
						if rec := recover(); rec != nil {
							_, _ = fmt.Fprintf(os.Stderr, "recovering panic calling function.\n%v\n", rec)
						}
					}()
					return fct(ctx, tck)
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
	}(ctx, o.d, o.f)

	return nil
}

func (o *run) checkMe() error {
	if o == nil {
		return ErrInvalid
	}

	o.m.RLock()
	defer o.m.RUnlock()

	if o.f == nil || o.d == 0 {
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
