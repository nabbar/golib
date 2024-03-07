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

package monitor

import (
	"context"
	"time"

	montps "github.com/nabbar/golib/monitor/types"
	librun "github.com/nabbar/golib/server/runner/ticker"
)

const (
	MaxPoolStart  = 15 * time.Second
	MaxTickPooler = 50 * time.Millisecond
)

func (o *mon) Start(ctx context.Context) error {
	if o.IsRunning() {
		_ = o.Stop(ctx)
	}

	if o == nil {
		return ErrorInvalid.Error(nil)
	}

	o.setRunner(ctx)

	o.m.RLock()
	defer o.m.RUnlock()
	if o.r == nil {
		return ErrorInvalid.Error(nil)
	}

	if e := o.r.Start(ctx); e != nil {
		return e
	} else {
		return o.poolIsRunning(ctx)
	}
}

func (o *mon) Stop(ctx context.Context) error {
	defer o.delRunner(ctx)

	if o == nil {
		return ErrorInvalid.Error(nil)
	} else if !o.IsRunning() {
		return nil
	}

	o.m.RLock()
	defer o.m.RUnlock()

	if o.r == nil {
		return ErrorInvalid.Error(nil)
	} else if e := o.r.Stop(ctx); e != nil {
		return e
	}

	return nil
}

func (o *mon) Restart(ctx context.Context) error {
	if e := o.Stop(ctx); e != nil {
		return e
	} else if e = o.Start(ctx); e != nil {
		return e
	}

	return nil
}

func (o *mon) IsRunning() bool {
	if o == nil {
		return false
	}

	o.m.RLock()
	defer o.m.RUnlock()

	if o.r == nil {
		return false
	} else {
		return o.r.IsRunning()
	}
}

func (o *mon) poolIsRunning(ctx context.Context) error {
	if o == nil {
		return ErrorInvalid.Error(nil)
	} else if o.IsRunning() {
		return nil
	}

	var (
		tck = time.NewTicker(MaxTickPooler)
		tms = time.Now()
	)

	defer tck.Stop()

	for {
		select {
		case <-tck.C:
			if o.IsRunning() {
				return nil
			} else if time.Since(tms) >= MaxPoolStart {
				return ErrorTimeout.Error(nil)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (o *mon) setRunner(ctx context.Context) {
	if o == nil {
		return
	}

	o.m.Lock()
	defer o.m.Unlock()

	if o.r != nil {
		_ = o.r.Stop(ctx)
	}

	var (
		cfg = o.getCfg()
	)

	o.r = librun.New(cfg.intervalCheck, o.runFunc)
}

func (o *mon) delRunner(ctx context.Context) {
	if o == nil {
		return
	}

	o.m.Lock()
	defer o.m.Unlock()

	if o.r != nil {
		_ = o.r.Stop(ctx)
	}

	o.r = nil
}

func (o *mon) runFunc(ctx context.Context, tck *time.Ticker) error {
	var (
		cfg = o.getCfg()
		chg = false
	)

	o.check(ctx, cfg)

	if o.IsRise() {
		tck.Reset(cfg.intervalRise)
		chg = true
	} else if o.IsFall() {
		tck.Reset(cfg.intervalFall)
		chg = true
	} else if chg {
		tck.Reset(cfg.intervalCheck)
		chg = false
	}

	return nil
}

func (o *mon) check(ctx context.Context, cfg *runCfg) {
	var fct montps.HealthCheck

	if fct = o.getFct(); fct == nil {
		l := o.getLastCheck()
		l.setStatus(ErrorMissingHealthCheck.Error(nil), 0, cfg)
		o.x.Store(keyLastRun, l)
	} else if cfg == nil {
		l := o.getLastCheck()
		l.setStatus(ErrorValidatorError.Error(nil), 0, cfg)
		o.x.Store(keyLastRun, l)
	}

	m := newMiddleware(cfg, fct)
	m.Add(o.mdlStatus)
	m.Run(ctx)

	// store metrics to prometheus exporter
	o.collectMetrics(ctx)
}
