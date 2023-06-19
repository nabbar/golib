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

package pool

import (
	"context"
	"sync"
	"time"

	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
	libprm "github.com/nabbar/golib/prometheus"
)

type pool struct {
	m  sync.RWMutex
	fp libprm.FuncGetPrometheus
	fl liblog.FuncLog
	p  libctx.Config[string]
}

func (o *pool) setDefaultLog() {
	o.m.Lock()
	defer o.m.Unlock()

	lg := liblog.New(o.p.GetContext)
	o.fl = func() liblog.Logger {
		return lg
	}
}

func (o *pool) getLog() liblog.Logger {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.fl != nil {
		if l := o.fl(); l != nil {
			return l
		}
	}

	o.m.RUnlock()
	o.setDefaultLog()
	o.m.RLock()

	return o.fl()
}

func (o *pool) InitMetrics(prm libprm.FuncGetPrometheus, log liblog.FuncLog) error {
	o.RegisterFctProm(prm)
	o.RegisterFctLogger(log)
	return o.createMetrics()
}

func (o *pool) RegisterFctProm(prm libprm.FuncGetPrometheus) {
	o.m.Lock()
	defer o.m.Unlock()

	o.fp = prm
}

func (o *pool) RegisterFctLogger(log liblog.FuncLog) {
	o.m.Lock()
	defer o.m.Unlock()

	o.fl = log
}

func (o *pool) TriggerCollectMetrics(ctx context.Context, dur time.Duration) {
	var tck *time.Ticker

	tck = time.NewTicker(dur)
	defer tck.Stop()

	for {
		select {
		case <-tck.C:
			if p := o.getProm(); p != nil {
				p.CollectMetrics(ctx)
			}

		case <-ctx.Done():
			return
		}
	}
}
