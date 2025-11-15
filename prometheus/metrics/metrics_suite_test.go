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

package metrics_test

import (
	"context"
	"testing"
	"time"

	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

var (
	// Global test context with timeout
	x  context.Context
	n  context.CancelFunc
	fx = func() context.Context {
		return x
	}
	// Global registerer for tests
	testRegistry prmsdk.Registerer
)

// TestMetrics is the entry point for Ginkgo test suite
func TestMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prometheus Metrics Suite")
}

var _ = BeforeSuite(func() {
	x, n = context.WithTimeout(context.Background(), 30*time.Second)
	testRegistry = prmsdk.NewRegistry()
})

var _ = AfterSuite(func() {
	if n != nil {
		n()
	}
})

// Helper function to create a new metric with proper registration
func newMetricWithRegistration(name string, metricType prmtps.MetricType) prmmet.Metric {
	m := prmmet.NewMetrics(name, metricType)
	Expect(m).ToNot(BeNil())
	return m
}

// Helper function to register a metric properly based on its type
func registerMetric(m prmmet.Metric) error {
	vec, err := m.GetType().Register(m)
	if err != nil {
		return err
	}
	return m.Register(testRegistry, vec)
}

// Helper function to unregister and cleanup a metric
func cleanupMetric(m prmmet.Metric) {
	if m != nil {
		Expect(m.UnRegister(testRegistry)).ToNot(HaveOccurred())
	}
}

// Helper function to create a gauge metric with default labels
func newGaugeMetric(name string, labels ...string) prmmet.Metric {
	m := newMetricWithRegistration(name, prmtps.Gauge)
	m.SetDesc("Test gauge metric")
	if len(labels) > 0 {
		m.AddLabel(labels...)
	}
	return m
}

// Helper function to create a counter metric with default labels
func newCounterMetric(name string, labels ...string) prmmet.Metric {
	m := newMetricWithRegistration(name, prmtps.Counter)
	m.SetDesc("Test counter metric")
	if len(labels) > 0 {
		m.AddLabel(labels...)
	}
	return m
}

// Helper function to create a histogram metric with buckets and labels
func newHistogramMetric(name string, buckets []float64, labels ...string) prmmet.Metric {
	m := newMetricWithRegistration(name, prmtps.Histogram)
	m.SetDesc("Test histogram metric")
	if len(buckets) > 0 {
		m.AddBuckets(buckets...)
	} else {
		m.AddBuckets(prmsdk.DefBuckets...)
	}
	if len(labels) > 0 {
		m.AddLabel(labels...)
	}
	return m
}

// Helper function to create a summary metric with objectives and labels
func newSummaryMetric(name string, objectives map[float64]float64, labels ...string) prmmet.Metric {
	m := newMetricWithRegistration(name, prmtps.Summary)
	m.SetDesc("Test summary metric")
	if len(objectives) > 0 {
		for k, v := range objectives {
			m.AddObjective(k, v)
		}
	} else {
		// Default objectives
		m.AddObjective(0.5, 0.05)
		m.AddObjective(0.9, 0.01)
		m.AddObjective(0.99, 0.001)
	}
	if len(labels) > 0 {
		m.AddLabel(labels...)
	}
	return m
}
