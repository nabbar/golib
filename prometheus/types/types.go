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

package types

import (
	"fmt"

	prmsdk "github.com/prometheus/client_golang/prometheus"
)

type MetricType int

const (
	None MetricType = iota
	Counter
	Gauge
	Histogram
	Summary
)

func (m MetricType) Register(met Metric) (prmsdk.Collector, error) {
	switch met.GetType() {
	case Counter:
		return m.counterHandler(met)
	case Gauge:
		return m.gaugeHandler(met)
	case Histogram:
		return m.histogramHandler(met)
	case Summary:
		return m.summaryHandler(met)
	}

	return nil, fmt.Errorf("metric type is not compatible")
}

func (m MetricType) counterHandler(met Metric) (prmsdk.Collector, error) {
	return prmsdk.NewCounterVec(
		prmsdk.CounterOpts{Name: met.GetName(), Help: met.GetDesc()},
		met.GetLabel(),
	), nil
}

func (m MetricType) gaugeHandler(met Metric) (prmsdk.Collector, error) {
	return prmsdk.NewGaugeVec(
		prmsdk.GaugeOpts{Name: met.GetName(), Help: met.GetDesc()},
		met.GetLabel(),
	), nil
}

func (m MetricType) histogramHandler(met Metric) (prmsdk.Collector, error) {
	if len(met.GetBuckets()) == 0 {
		return nil, fmt.Errorf("metric '%s' is histogram type, cannot lose bucket param", met.GetName())
	}

	return prmsdk.NewHistogramVec(
		prmsdk.HistogramOpts{Name: met.GetName(), Help: met.GetDesc(), Buckets: met.GetBuckets()},
		met.GetLabel(),
	), nil
}

func (m MetricType) summaryHandler(met Metric) (prmsdk.Collector, error) {
	if len(met.GetObjectives()) == 0 {
		return nil, fmt.Errorf("metric '%s' is summary type, cannot lose objectives param", met.GetName())
	}

	return prmsdk.NewSummaryVec(
		prmsdk.SummaryOpts{Name: met.GetName(), Help: met.GetDesc(), Objectives: met.GetObjectives()},
		met.GetLabel(),
	), nil
}
