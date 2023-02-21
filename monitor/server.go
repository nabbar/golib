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

	"github.com/nabbar/golib/monitor/types"
)

func (o *mon) setChan() {
	o.m.Lock()
	defer o.m.Unlock()

	o.s = make(chan struct{})
}

func (o *mon) getChan() <-chan struct{} {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.s
}

func (o *mon) sendChan() {
	o.m.RLock()
	defer o.m.RUnlock()

	o.s <- struct{}{}
}

func (o *mon) Start(ctx context.Context) error {
	if o.IsRunning() {
		_ = o.Stop(ctx)
	}

	o.setChan()
	go o.ticker(ctx)

	if o.IsRunning() {
		return nil
	}

	t := time.Now()
	for {
		if time.Since(t) > 15*time.Second {
			return ErrorTimeout.Error(nil)
		} else if o.IsRunning() {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (o *mon) Stop(ctx context.Context) error {
	if !o.IsRunning() {
		return nil
	}

	t := time.Now()

	for {
		o.sendChan()
		time.Sleep(100 * time.Millisecond)

		if time.Since(t) > 15*time.Second {
			return ErrorTimeout.Error(nil)
		} else if !o.IsRunning() {
			return nil
		}
	}
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
	if i, l := o.x.Load(keyRun); !l {
		return false
	} else if v, k := i.(bool); !k {
		return false
	} else {
		return v
	}
}

func (o *mon) setRunning(state bool) {
	if state {
		o.x.Store(keyRun, state)
	} else {
		o.x.Delete(keyRun)
	}
}

func (o *mon) ticker(ctx context.Context) {
	var (
		cfg = o.getCfg()
		chg = false
		tck *time.Ticker
	)

	tck = time.NewTicker(cfg.intervalCheck)
	defer tck.Stop()

	o.setRunning(true)
	defer o.setRunning(false)

	for {
		select {
		case <-tck.C:
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

		case <-ctx.Done():
			return

		case <-o.getChan():
			return
		}
	}
}

func (o *mon) check(ctx context.Context, cfg *runCfg) {
	var fct types.HealthCheck

	if fct = o.getFct(); fct == nil {
		_ = o.setStatus(ErrorMissingHealthCheck.Error(nil), cfg)
	} else if cfg == nil {
		_ = o.setStatus(ErrorValidatorError.Error(nil), cfg)
	}

	m := newMiddleware(cfg, fct)
	m.Add(o.setLatency)
	m.Add(o.mdlStatus)
	// add here other part to run

	m.Add(o.setUpTime)
	m.Add(o.setDownTime)
	m.Add(o.setRiseTime)
	m.Add(o.setFallTime)

	// no add after this
	m.Add(o.setLastCheck)
	m.Run(ctx)

	// store metrics to prometheus exporter
	o.collectMetrics(ctx)
}
