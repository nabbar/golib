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
	"strings"

	montps "github.com/nabbar/golib/monitor/types"
	libprm "github.com/nabbar/golib/prometheus"
	libmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
)

const (
	metricBaseName  = "monitor"
	metricLatency   = "latency"
	metricUptime    = "uptime"
	metricDowntime  = "downtime"
	metricRisetime  = "risetime"
	metricFalltime  = "falltime"
	metricStatus    = "status"
	metricBoolTrue  = "true"
	metricBoolFalse = "false"
)

func (o *pool) normalizeName(name string) string {
	name = strings.ToLower(name)

	name = strings.Replace(name, " ", "_", -1)
	name = strings.Replace(name, "-", "_", -1)
	name = strings.Replace(name, ".", "", -1)

	for strings.Contains(name, "__") {
		name = strings.Replace(name, "__", "_", -1)
	}

	return name
}

func (o *pool) getMetricName(monitor, metric string) string {
	part := make([]string, 0)
	part = append(part, o.normalizeName(metricBaseName))
	part = append(part, o.normalizeName(metric))
	part = append(part, o.normalizeName(monitor))
	return strings.Join(part, "_")
}

func (o *pool) getProm() libprm.Prometheus {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.fp == nil {
		return nil
	} else if p := o.fp(); p != nil {
		return p
	}

	return nil
}

func (o *pool) createMetrics(mon montps.Monitor) error {
	var prm libprm.Prometheus

	if prm = o.getProm(); prm == nil {
		return nil
	}

	var (
		name = mon.Name()
		met  libmet.Metric
		mnm  string
	)

	mnm = o.getMetricName(name, metricLatency)
	mon.RegisterMetricsAddName(mnm)
	met = libmet.NewMetrics(mnm, prmtps.Histogram)
	met.SetDesc("the time monitor took to check the health of '" + name + "' component")
	met.AddLabel(metricBaseName)
	met.AddBuckets(prm.GetDuration()...)
	met.SetCollect(func(ctx context.Context, m libmet.Metric) {
		_ = m.Observe([]string{name}, mon.CollectLatency().Seconds())
	})

	if e := prm.AddMetric(false, met); e != nil {
		return e
	}

	mnm = o.getMetricName(name, metricUptime)
	mon.RegisterMetricsAddName(mnm)
	met = libmet.NewMetrics(mnm, prmtps.Histogram)
	met.SetDesc("the total time during which the '" + name + "' component is up")
	met.AddLabel(metricBaseName)
	met.AddBuckets(prm.GetDuration()...)
	met.SetCollect(func(ctx context.Context, m libmet.Metric) {
		_ = m.Observe([]string{name}, mon.CollectUpTime().Seconds())
	})

	if e := prm.AddMetric(false, met); e != nil {
		return e
	}

	mnm = o.getMetricName(name, metricDowntime)
	mon.RegisterMetricsAddName(mnm)
	met = libmet.NewMetrics(mnm, prmtps.Histogram)
	met.SetDesc("the total time during which the '" + name + "' component is down")
	met.AddLabel(metricBaseName)
	met.AddBuckets(prm.GetDuration()...)
	met.SetCollect(func(ctx context.Context, m libmet.Metric) {
		_ = m.Observe([]string{name}, mon.CollectDownTime().Seconds())
	})

	if e := prm.AddMetric(false, met); e != nil {
		return e
	}

	mnm = o.getMetricName(name, metricRisetime)
	mon.RegisterMetricsAddName(mnm)
	met = libmet.NewMetrics(mnm, prmtps.Histogram)
	met.SetDesc("the total time during which the '" + name + "' component is rising")
	met.AddLabel(metricBaseName)
	met.AddBuckets(prm.GetDuration()...)
	met.SetCollect(func(ctx context.Context, m libmet.Metric) {
		_ = m.Observe([]string{name}, mon.CollectRiseTime().Seconds())
	})

	if e := prm.AddMetric(false, met); e != nil {
		return e
	}

	mnm = o.getMetricName(name, metricFalltime)
	mon.RegisterMetricsAddName(mnm)
	met = libmet.NewMetrics(mnm, prmtps.Histogram)
	met.SetDesc("the total time during which the '" + name + "' component is falling")
	met.AddLabel(metricBaseName)
	met.AddBuckets(prm.GetDuration()...)
	met.SetCollect(func(ctx context.Context, m libmet.Metric) {
		_ = m.Observe([]string{name}, mon.CollectFallTime().Seconds())
	})

	if e := prm.AddMetric(false, met); e != nil {
		return e
	}

	mnm = o.getMetricName(name, metricStatus)
	mon.RegisterMetricsAddName(mnm)
	met = libmet.NewMetrics(mnm, prmtps.Counter)
	met.SetDesc("the total time during which the '" + name + "' component is falling")
	met.AddLabel(metricBaseName, "status", "rise", "fall")
	met.SetCollect(func(ctx context.Context, m libmet.Metric) {
		var (
			s, r, f = mon.CollectStatus()
			rs, fs  string
		)

		if r {
			rs = metricBoolTrue
		} else {
			rs = metricBoolFalse
		}

		if f {
			fs = metricBoolTrue
		} else {
			fs = metricBoolFalse
		}

		_ = m.Inc([]string{name, s, rs, fs})
	})

	if e := prm.AddMetric(false, met); e != nil {
		return e
	}

	mon.RegisterCollectMetrics(prm.CollectMetrics)

	return nil
}

func (o *pool) deleteMetrics(mon montps.Monitor) {
	var prm libprm.Prometheus

	if prm = o.getProm(); prm == nil {
		return
	}

	var (
		name = mon.Name()
	)

	prm.DelMetric(o.getMetricName(name, metricLatency))
	prm.DelMetric(o.getMetricName(name, metricUptime))
	prm.DelMetric(o.getMetricName(name, metricDowntime))
	prm.DelMetric(o.getMetricName(name, metricRisetime))
	prm.DelMetric(o.getMetricName(name, metricFalltime))
	prm.DelMetric(o.getMetricName(name, metricStatus))
}
