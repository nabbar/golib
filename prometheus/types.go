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
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricType int

const (
	None MetricType = iota
	Counter
	Gauge
	Histogram
	Summary
)

func (m MetricType) Register(metric *metrics) error {
	switch metric.t {
	case Counter:
		return m.counterHandler(metric)
	case Gauge:
		return m.gaugeHandler(metric)
	case Histogram:
		return m.histogramHandler(metric)
	case Summary:
		return m.summaryHandler(metric)
	}

	return errors.Errorf("metric type is not compatible.")
}

func (m MetricType) counterHandler(metric *metrics) error {
	metric.v = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: metric.n, Help: metric.d},
		metric.l,
	)
	return nil
}

func (m MetricType) gaugeHandler(metric *metrics) error {
	metric.v = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: metric.n, Help: metric.d},
		metric.l,
	)
	return nil
}

func (m MetricType) histogramHandler(metric *metrics) error {
	if len(metric.b) == 0 {
		return errors.Errorf("metric '%s' is histogram type, cannot lose bucket param.", metric.n)
	}
	metric.v = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: metric.n, Help: metric.d, Buckets: metric.b},
		metric.l,
	)
	return nil
}

func (m MetricType) summaryHandler(metric *metrics) error {
	if len(metric.o) == 0 {
		return errors.Errorf("metric '%s' is summary type, cannot lose objectives param.", metric.n)
	}
	prometheus.NewSummaryVec(
		prometheus.SummaryOpts{Name: metric.n, Help: metric.d, Objectives: metric.o},
		metric.l,
	)
	return nil
}
