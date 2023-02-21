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

type FuncGetMetric func() Metric
type FuncGetMetricName func(name string) Metric
type FuncCollectMetrics func() Collect
type FuncCollectMetricsByName func(name string) Collect

type Collect interface {
	Collect(ctx context.Context)
}

type Metric interface {
	prmtps.Metric
	Collect

	SetGaugeValue(labelValues []string, value float64) error
	Inc(labelValues []string) error
	Add(labelValues []string, value float64) error
	Observe(labelValues []string, value float64) error

	Register(vec prmsdk.Collector) error
	UnRegister() bool

	SetDesc(desc string)
	AddLabel(label ...string)
	AddBuckets(bucket ...float64)
	AddObjective(key, value float64)

	SetCollect(fct FuncCollect)
	GetCollect() FuncCollect
}

type FuncCollect func(ctx context.Context, m Metric)

func NewMetrics(name string, metricType prmtps.MetricType) Metric {
	return &metrics{
		m: sync.RWMutex{},
		t: metricType,
		n: name,
		d: "",
		l: make([]string, 0),
		b: make([]float64, 0),
		o: make(map[float64]float64, 0),
		f: nil,
		v: nil,
	}
}
