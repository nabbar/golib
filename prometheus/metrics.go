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
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics interface {
	SetGaugeValue(labelValues []string, value float64) error
	Inc(labelValues []string) error
	Add(labelValues []string, value float64) error
	Observe(labelValues []string, value float64) error

	Collect(c *gin.Context, start time.Time)

	GetName() string
	GetType() MetricType
	SetDesc(desc string)

	AddLabel(label ...string)
	AddBuckets(bucket ...float64)
	AddObjective(key, value float64)
	SetCollect(fct FuncCollect)
}

type FuncCollect func(m Metrics, c *gin.Context, start time.Time)

func NewMetrics(name string, metricType MetricType) Metrics {
	return &metrics{
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

// metrics defines a metric object. Users can use it to save
// metric data. Every metric should be globally unique by name.
type metrics struct {
	m sync.Mutex
	t MetricType
	n string
	d string
	l []string
	b []float64
	o map[float64]float64
	f FuncCollect
	v prometheus.Collector
}

// SetGaugeValue set data for Gauge type metrics.
func (m *metrics) SetGaugeValue(labelValues []string, value float64) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.t == None {
		return errors.Errorf("metric '%s' not existed.", m.n)
	}

	if m.t != Gauge {
		return errors.Errorf("metric '%s' not Gauge type", m.n)
	}

	m.v.(*prometheus.GaugeVec).WithLabelValues(labelValues...).Set(value)
	return nil
}

// Inc increases value for Counter/Gauge type metric, increments
// the counter by 1
func (m *metrics) Inc(labelValues []string) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.t == None {
		return errors.Errorf("metric '%s' not existed.", m.n)
	}

	if m.t != Gauge && m.t != Counter {
		return errors.Errorf("metric '%s' not Gauge or Counter type", m.n)
	}

	switch m.t {
	case Counter:
		m.v.(*prometheus.CounterVec).WithLabelValues(labelValues...).Inc()
		break
	case Gauge:
		m.v.(*prometheus.GaugeVec).WithLabelValues(labelValues...).Inc()
		break
	}

	return nil
}

// Add adds the given value to the metrics object. Only
// for Counter/Gauge type metric.
func (m *metrics) Add(labelValues []string, value float64) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.t == None {
		return errors.Errorf("metric '%s' not existed.", m.n)
	}

	if m.t != Gauge && m.t != Counter {
		return errors.Errorf("metric '%s' not Gauge or Counter type", m.n)
	}

	switch m.t {
	case Counter:
		m.v.(*prometheus.CounterVec).WithLabelValues(labelValues...).Add(value)
		break
	case Gauge:
		m.v.(*prometheus.GaugeVec).WithLabelValues(labelValues...).Add(value)
		break
	}

	return nil
}

// Observe is used by Histogram and Summary type metric to
// add observations.
func (m *metrics) Observe(labelValues []string, value float64) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.t == 0 {
		return errors.Errorf("metric '%s' not existed.", m.n)
	}

	if m.t != Histogram && m.t != Summary {
		return errors.Errorf("metric '%s' not Histogram or Summary type", m.n)
	}

	switch m.t {
	case Histogram:
		m.v.(*prometheus.HistogramVec).WithLabelValues(labelValues...).Observe(value)
		break
	case Summary:
		m.v.(*prometheus.SummaryVec).WithLabelValues(labelValues...).Observe(value)
		break
	}

	return nil
}

func (m *metrics) Collect(c *gin.Context, start time.Time) {
	if m.f != nil {
		m.f(m, c, start)
	}
}

func (m *metrics) GetName() string {
	return m.n
}

func (m *metrics) GetType() MetricType {
	return m.t
}

func (m *metrics) SetDesc(desc string) {
	m.m.Lock()
	defer m.m.Unlock()

	m.d = desc
}

func (m *metrics) AddLabel(label ...string) {
	m.m.Lock()
	defer m.m.Unlock()

	if len(m.l) < 1 {
		m.l = make([]string, 0)
	}

	m.l = append(m.l, label...)
}

func (m *metrics) AddBuckets(bucket ...float64) {
	m.m.Lock()
	defer m.m.Unlock()

	if len(m.b) < 1 {
		m.b = make([]float64, 0)
	}

	m.b = append(m.b, bucket...)
}

func (m *metrics) AddObjective(key, value float64) {
	m.m.Lock()
	defer m.m.Unlock()

	if len(m.o) < 1 {
		m.o = make(map[float64]float64, 0)
	}

	m.o[key] = value
}

func (m *metrics) SetCollect(fct FuncCollect) {
	m.m.Lock()
	defer m.m.Unlock()

	m.f = fct
}
