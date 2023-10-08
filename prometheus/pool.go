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

package prometheus

import (
	"context"

	ginsdk "github.com/gin-gonic/gin"
	libmet "github.com/nabbar/golib/prometheus/metrics"
	prmpol "github.com/nabbar/golib/prometheus/pool"
	libsem "github.com/nabbar/golib/semaphore"
)

// GetMetric used to retrieve a metric instance by his name.
func (m *prom) GetMetric(name string) libmet.Metric {
	if v := m.ginMet.Get(name); v != nil {
		return v
	} else if v = m.othMet.Get(name); v != nil {
		return v
	}

	return nil
}

// AddMetric used to store the metric object to the API Metric Pool or Other Metric Pool.
func (m *prom) AddMetric(isAPI bool, metric libmet.Metric) error {
	if isAPI {
		return m.ginMet.Add(metric)
	} else {
		return m.othMet.Add(metric)
	}
}

// DelMetric used to unregister the metric and delete the instance from Pool.
func (m *prom) DelMetric(name string) {
	m.ginMet.Del(name)
	m.othMet.Del(name)
}

// ListMetric retrieve a slice of metrics' name registered for all type, API or not.
func (m *prom) ListMetric() []string {
	var res = make([]string, 0)
	res = append(res, m.ginMet.List()...)
	res = append(res, m.othMet.List()...)
	return res
}

func (m *prom) Collect(ctx context.Context) {
	m.CollectMetrics(ctx)
}

func (m *prom) CollectMetrics(ctx context.Context, name ...string) {
	var (
		ok bool
		s  libsem.Semaphore
	)

	s = libsem.New(ctx, 0, false)
	defer s.DeferMain()

	if _, ok = ctx.(*ginsdk.Context); ok {
		m.ginMet.Walk(m.runCollect(ctx, s), name...)
	} else {
		m.othMet.Walk(m.runCollect(ctx, s), name...)
	}

	_ = s.WaitAll()
}

func (m *prom) runCollect(ctx context.Context, sem libsem.Semaphore) prmpol.FuncWalk {
	return func(pool prmpol.MetricPool, key string, val libmet.Metric) bool {
		if e := sem.NewWorker(); e != nil {
			return false
		}

		go func() {
			defer sem.DeferWorker()
			val.Collect(ctx)
			pool.Set(key, val)
		}()

		return true
	}
}
