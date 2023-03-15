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

	liblog "github.com/nabbar/golib/logger"

	montps "github.com/nabbar/golib/monitor/types"
	libprm "github.com/nabbar/golib/prometheus"
	libmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
)

const (
	metricBaseName = "monitor"
	metricLatency  = "latency"
	metricUptime   = "uptime"
	metricDowntime = "downtime"
	metricRiseTime = "risetime"
	metricFallTime = "falltime"
	metricStatus   = "status"
	metricRise     = "rise"
	metricFall     = "fall"
	metricSLis     = "sli"

	monitorMeans = "mean"
	monitorMin   = "min"
	monitorMax   = "max"
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

func (o *pool) getMetricName(metric string) string {
	part := make([]string, 0)
	part = append(part, o.normalizeName(metricBaseName))
	part = append(part, o.normalizeName(metric))
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

func (o *pool) createMetricsLatency() error {
	var (
		prm libprm.Prometheus
		met libmet.Metric
		mnm string
	)

	if prm = o.getProm(); prm == nil {
		return nil
	}

	mnm = o.getMetricName(metricLatency)
	met = libmet.NewMetrics(mnm, prmtps.Histogram)
	met.SetDesc("the time monitor took to check components' health")
	met.AddLabel(metricBaseName)
	met.AddBuckets([]float64{0.0, 0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.75, 1.0, 1.5, 2.0, 2.5, 3.0, 4.0, 5.0, 10}...)
	met.SetCollect(o.collectMetricLatency)

	return prm.AddMetric(false, met)
}

func (o *pool) collectMetricLatency(ctx context.Context, m libmet.Metric) {
	var log = o.getLog()

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if e := m.Observe([]string{name}, val.CollectLatency().Seconds()); e != nil {
			ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
			ent.FieldAdd("monitor", name)
			ent.FieldAdd("metric", val.Name())
			ent.ErrorAdd(true, e)
			ent.Log()
		}

		return true
	})
}

func (o *pool) createMetricsUptime() error {
	var (
		prm libprm.Prometheus
		met libmet.Metric
		mnm string
	)

	if prm = o.getProm(); prm == nil {
		return nil
	}

	mnm = o.getMetricName(metricUptime)
	met = libmet.NewMetrics(mnm, prmtps.Gauge)
	met.SetDesc("the total seconds during the component are up")
	met.AddLabel(metricBaseName)
	met.SetCollect(o.collectMetricUptime)

	return prm.AddMetric(false, met)
}

func (o *pool) collectMetricUptime(ctx context.Context, m libmet.Metric) {
	var log = o.getLog()

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if e := m.SetGaugeValue([]string{name}, val.CollectUpTime().Seconds()); e != nil {
			ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
			ent.FieldAdd("monitor", name)
			ent.FieldAdd("metric", val.Name())
			ent.ErrorAdd(true, e)
			ent.Log()
		}

		return true
	})
}

func (o *pool) createMetricsDowntime() error {
	var (
		prm libprm.Prometheus
		met libmet.Metric
		mnm string
	)

	if prm = o.getProm(); prm == nil {
		return nil
	}

	mnm = o.getMetricName(metricDowntime)
	met = libmet.NewMetrics(mnm, prmtps.Gauge)
	met.SetDesc("the total time during the components are down")
	met.AddLabel(metricBaseName)
	met.SetCollect(o.collectMetricDowntime)

	return prm.AddMetric(false, met)
}

func (o *pool) collectMetricDowntime(ctx context.Context, m libmet.Metric) {
	var log = o.getLog()

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if e := m.SetGaugeValue([]string{name}, val.CollectDownTime().Seconds()); e != nil {
			ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
			ent.FieldAdd("monitor", name)
			ent.FieldAdd("metric", val.Name())
			ent.ErrorAdd(true, e)
			ent.Log()
		}

		return true
	})
}

func (o *pool) createMetricsRiseTime() error {
	var (
		prm libprm.Prometheus
		met libmet.Metric
		mnm string
	)

	if prm = o.getProm(); prm == nil {
		return nil
	}

	mnm = o.getMetricName(metricRiseTime)
	met = libmet.NewMetrics(mnm, prmtps.Gauge)
	met.SetDesc("the total time during the components are rising")
	met.AddLabel(metricBaseName)
	met.SetCollect(o.collectMetricRiseTime)

	return prm.AddMetric(false, met)
}

func (o *pool) collectMetricRiseTime(ctx context.Context, m libmet.Metric) {
	var log = o.getLog()

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if e := m.SetGaugeValue([]string{name}, val.CollectRiseTime().Seconds()); e != nil {
			ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
			ent.FieldAdd("monitor", name)
			ent.FieldAdd("metric", val.Name())
			ent.ErrorAdd(true, e)
			ent.Log()
		}

		return true
	})
}

func (o *pool) createMetricsFallTime() error {
	var (
		prm libprm.Prometheus
		met libmet.Metric
		mnm string
	)

	if prm = o.getProm(); prm == nil {
		return nil
	}

	mnm = o.getMetricName(metricFallTime)
	met = libmet.NewMetrics(mnm, prmtps.Gauge)
	met.SetDesc("the total time during the components are falling")
	met.AddLabel(metricBaseName)
	met.SetCollect(o.collectMetricFallTime)

	return prm.AddMetric(false, met)
}

func (o *pool) collectMetricFallTime(ctx context.Context, m libmet.Metric) {
	var log = o.getLog()

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if e := m.SetGaugeValue([]string{name}, val.CollectFallTime().Seconds()); e != nil {
			ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
			ent.FieldAdd("monitor", name)
			ent.FieldAdd("metric", val.Name())
			ent.ErrorAdd(true, e)
			ent.Log()
		}

		return true
	})
}

func (o *pool) createMetricsStatus() error {
	var (
		prm libprm.Prometheus
		met libmet.Metric
		mnm string
	)

	if prm = o.getProm(); prm == nil {
		return nil
	}

	mnm = o.getMetricName(metricStatus)
	met = libmet.NewMetrics(mnm, prmtps.Gauge)
	met.SetDesc("the instant status of components")
	met.AddLabel(metricBaseName)
	met.SetCollect(o.collectMetricStatus)

	return prm.AddMetric(false, met)
}

func (o *pool) collectMetricStatus(ctx context.Context, m libmet.Metric) {
	var log = o.getLog()

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		var (
			s, _, _ = val.CollectStatus()
		)

		if e := m.SetGaugeValue([]string{name}, s.Float()); e != nil {
			ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
			ent.FieldAdd("monitor", name)
			ent.FieldAdd("metric", val.Name())
			ent.ErrorAdd(true, e)
			ent.Log()
		}
		return true
	})
}

func (o *pool) createMetricsRising() error {
	var (
		prm libprm.Prometheus
		met libmet.Metric
		mnm string
	)

	if prm = o.getProm(); prm == nil {
		return nil
	}

	mnm = o.getMetricName(metricRise)
	met = libmet.NewMetrics(mnm, prmtps.Gauge)
	met.SetDesc("the status rising value of components")
	met.AddLabel(metricBaseName)
	met.SetCollect(o.collectMetricRising)

	return prm.AddMetric(false, met)
}

func (o *pool) collectMetricRising(ctx context.Context, m libmet.Metric) {
	var log = o.getLog()

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		var (
			_, r, _ = val.CollectStatus()
			s       float64
		)

		if r {
			s = 1
		} else {
			s = 0
		}

		if e := m.SetGaugeValue([]string{name}, s); e != nil {
			ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
			ent.FieldAdd("monitor", name)
			ent.FieldAdd("metric", val.Name())
			ent.ErrorAdd(true, e)
			ent.Log()
		}

		return true
	})
}

func (o *pool) createMetricsFalling() error {
	var (
		prm libprm.Prometheus
		met libmet.Metric
		mnm string
	)

	if prm = o.getProm(); prm == nil {
		return nil
	}

	mnm = o.getMetricName(metricFall)
	met = libmet.NewMetrics(mnm, prmtps.Gauge)
	met.SetDesc("the status falling value of components")
	met.AddLabel(metricBaseName)
	met.SetCollect(o.collectMetricFalling)

	return prm.AddMetric(false, met)
}

func (o *pool) collectMetricFalling(ctx context.Context, m libmet.Metric) {
	var log = o.getLog()

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		var (
			_, _, f = val.CollectStatus()
			s       float64
		)

		if f {
			s = 1
		} else {
			s = 0
		}

		if e := m.SetGaugeValue([]string{name}, s); e != nil {
			ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
			ent.FieldAdd("monitor", name)
			ent.FieldAdd("metric", val.Name())
			ent.ErrorAdd(true, e)
			ent.Log()
		}

		return true
	})
}

func (o *pool) createMetricsSLis() error {
	var (
		prm libprm.Prometheus
		met libmet.Metric
		mnm string
	)

	if prm = o.getProm(); prm == nil {
		return nil
	}

	mnm = o.getMetricName(metricSLis)
	met = libmet.NewMetrics(mnm, prmtps.Gauge)
	met.SetDesc("the SLIs rate of each component")
	met.AddLabel(metricBaseName)
	met.SetCollect(o.collectMetricSLis)

	return prm.AddMetric(false, met)
}

func (o *pool) collectMetricSLis(ctx context.Context, m libmet.Metric) {
	var (
		log         = o.getLog()
		min float64 = 1
		max float64 = 0
		cur float64 = 0
		cnt int     = 0
		sum float64 = 0
	)

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		cur = val.CollectDownTime().Seconds() / val.CollectUpTime().Seconds()
		sum += cur
		cnt++

		if cur < min {
			min = cur
		}
		if cur > max {
			max = cur
		}

		if e := m.SetGaugeValue([]string{name}, cur); e != nil {
			ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
			ent.FieldAdd("monitor", name)
			ent.FieldAdd("metric", val.Name())
			ent.ErrorAdd(true, e)
			ent.Log()
		}

		return true
	})

	mns := 1 - (sum / float64(cnt))
	min = 1 - min
	max = 1 - max

	if mns < 0 {
		mns = 0
	}
	if min < 0 {
		min = 0
	}
	if max < 0 {
		max = 0
	}

	if e := m.SetGaugeValue([]string{monitorMeans}, mns); e != nil {
		ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
		ent.FieldAdd("monitor", monitorMeans)
		ent.FieldAdd("metric", metricSLis)
		ent.ErrorAdd(true, e)
		ent.Log()
	}

	if e := m.SetGaugeValue([]string{monitorMin}, min); e != nil {
		ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
		ent.FieldAdd("monitor", monitorMin)
		ent.FieldAdd("metric", metricSLis)
		ent.ErrorAdd(true, e)
		ent.Log()
	}

	if e := m.SetGaugeValue([]string{monitorMax}, max); e != nil {
		ent := log.Entry(liblog.ErrorLevel, "failed to collect metrics", nil)
		ent.FieldAdd("monitor", monitorMax)
		ent.FieldAdd("metric", metricSLis)
		ent.ErrorAdd(true, e)
		ent.Log()
	}
}

func (o *pool) createMetrics() error {
	if e := o.createMetricsLatency(); e != nil {
		return e
	}

	if e := o.createMetricsUptime(); e != nil {
		return e
	}

	if e := o.createMetricsDowntime(); e != nil {
		return e
	}

	if e := o.createMetricsRiseTime(); e != nil {
		return e
	}

	if e := o.createMetricsFallTime(); e != nil {
		return e
	}

	if e := o.createMetricsStatus(); e != nil {
		return e
	}

	if e := o.createMetricsRising(); e != nil {
		return e
	}

	if e := o.createMetricsFalling(); e != nil {
		return e
	}

	if e := o.createMetricsSLis(); e != nil {
		return e
	}

	return nil
}
