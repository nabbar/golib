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
	"fmt"

	prmtps "github.com/nabbar/golib/prometheus/types"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// SetGaugeValue set data for Gauge type ginMet.
func (m *metrics) SetGaugeValue(labelValues []string, value float64) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.t == prmtps.None {
		return fmt.Errorf("metric '%s' not existed", m.n)
	}

	if m.t != prmtps.Gauge {
		return fmt.Errorf("metric '%s' not Gauge type", m.n)
	}

	m.v.(*prmsdk.GaugeVec).WithLabelValues(labelValues...).Set(value)
	return nil
}

// Inc increases value for Counter/Gauge type metric, increments
// the counter by 1
func (m *metrics) Inc(labelValues []string) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.t == prmtps.None {
		return fmt.Errorf("metric '%s' not existed", m.n)
	}

	if m.t != prmtps.Gauge && m.t != prmtps.Counter {
		return fmt.Errorf("metric '%s' not Gauge or Counter type", m.n)
	}

	switch m.t {
	case prmtps.Counter:
		m.v.(*prmsdk.CounterVec).WithLabelValues(labelValues...).Inc()
		break
	case prmtps.Gauge:
		m.v.(*prmsdk.GaugeVec).WithLabelValues(labelValues...).Inc()
		break
	}

	return nil
}

// Add adds the given value to the ginMet object. Only
// for Counter/Gauge type metric.
func (m *metrics) Add(labelValues []string, value float64) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.t == prmtps.None {
		return fmt.Errorf("metric '%s' not existed", m.n)
	}

	if m.t != prmtps.Gauge && m.t != prmtps.Counter {
		return fmt.Errorf("metric '%s' not Gauge or Counter type", m.n)
	}

	switch m.t {
	case prmtps.Counter:
		m.v.(*prmsdk.CounterVec).WithLabelValues(labelValues...).Add(value)
		break
	case prmtps.Gauge:
		m.v.(*prmsdk.GaugeVec).WithLabelValues(labelValues...).Add(value)
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
		return fmt.Errorf("metric '%s' not existed", m.n)
	}

	if m.t != prmtps.Histogram && m.t != prmtps.Summary {
		return fmt.Errorf("metric '%s' not Histogram or Summary type", m.n)
	}

	switch m.t {
	case prmtps.Histogram:
		m.v.(*prmsdk.HistogramVec).WithLabelValues(labelValues...).Observe(value)
		break
	case prmtps.Summary:
		m.v.(*prmsdk.SummaryVec).WithLabelValues(labelValues...).Observe(value)
		break
	}

	return nil
}
