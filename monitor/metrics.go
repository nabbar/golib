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

	monsts "github.com/nabbar/golib/monitor/status"
	libprm "github.com/nabbar/golib/prometheus"
	"golang.org/x/exp/slices"
)

func (o *mon) RegisterMetricsName(names ...string) {
	o.x.Store(keyMetricsName, names)
}

func (o *mon) RegisterMetricsAddName(names ...string) {
	var n []string
	if i, l := o.x.Load(keyMetricsName); !l || i == nil {
		n = make([]string, 0)
	} else if v, k := i.([]string); !k {
		n = make([]string, 0)
	} else {
		n = v
	}

	for _, i := range names {
		if !slices.Contains(n, i) {
			n = append(n, i)
		}
	}

	o.x.Store(keyMetricsName, n)
}

func (o *mon) RegisterCollectMetrics(fct libprm.FuncCollectMetrics) {
	o.x.Store(keyMetricsFunc, fct)
}

func (o *mon) CollectLatency() time.Duration {
	return o.Latency()
}

func (o *mon) CollectUpTime() time.Duration {
	return o.Uptime()
}

func (o *mon) CollectDownTime() time.Duration {
	return o.Downtime()
}

func (o *mon) CollectRiseTime() time.Duration {
	return o.getLastCheck().RiseTime()
}

func (o *mon) CollectFallTime() time.Duration {
	return o.getLastCheck().FallTime()
}

func (o *mon) CollectStatus() (sts monsts.Status, rise bool, fall bool) {
	return o.Status(), o.IsRise(), o.IsFall()
}

func (o *mon) collectMetrics(ctx context.Context) {
	var (
		n []string
		f libprm.FuncCollectMetrics
	)

	if i, l := o.x.Load(keyMetricsName); !l || i == nil {
		return
	} else if v, k := i.([]string); !k {
		return
	} else {
		n = v
	}

	if i, l := o.x.Load(keyMetricsFunc); !l {
		return
	} else if v, k := i.(libprm.FuncCollectMetrics); !k {
		return
	} else {
		f = v
	}

	if len(n) < 1 || f == nil {
		return
	}

	f(ctx, n...)
}
