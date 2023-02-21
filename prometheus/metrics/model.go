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

package metrics

import (
	"context"
	"sync"

	prmtps "github.com/nabbar/golib/prometheus/types"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// metrics defines a metric object. Users can use it to save
// metric data. Every metric should be globally unique by name.
type metrics struct {
	m sync.RWMutex
	t prmtps.MetricType
	n string
	d string
	l []string
	b []float64
	o map[float64]float64
	f FuncCollect
	v prmsdk.Collector
}

func (m *metrics) SetCollect(fct FuncCollect) {
	m.m.Lock()
	defer m.m.Unlock()

	m.f = fct
}

func (m *metrics) GetCollect() FuncCollect {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.f
}

func (m *metrics) GetName() string {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.n
}

func (m *metrics) GetType() prmtps.MetricType {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.t
}

func (m *metrics) Collect(ctx context.Context) {
	if m.f != nil {
		m.f(ctx, m)
	}
}

func (m *metrics) Register(vec prmsdk.Collector) error {
	m.m.Lock()
	defer m.m.Unlock()
	m.v = vec
	prmsdk.Unregister(m.v)
	return prmsdk.Register(m.v)
}

func (m *metrics) UnRegister() bool {
	m.m.Lock()
	defer m.m.Unlock()
	return prmsdk.Unregister(m.v)
}
