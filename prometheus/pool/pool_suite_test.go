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

package pool_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmpool "github.com/nabbar/golib/prometheus/pool"
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

var (
	// Global test context with timeout
	testCtx    context.Context
	cancelFunc context.CancelFunc

	// Global counter for unique metric names
	metricCounter atomic.Uint64

	// Global registerer for tests
	testRegistry prmsdk.Registerer
)

// TestPool is the entry point for Ginkgo test suite
func TestPool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prometheus Metric Pool Suite")
}

var _ = BeforeSuite(func() {
	testCtx, cancelFunc = context.WithTimeout(context.Background(), 60*time.Second)
	testRegistry = prmsdk.NewRegistry()
})

var _ = AfterSuite(func() {
	if cancelFunc != nil {
		cancelFunc()
	}
})

// Helper function to create a new pool
func newPool() prmpool.MetricPool {
	return prmpool.New(testCtx, prmsdk.NewRegistry())
}

// Helper function to create a counter metric
func createCounterMetric(name string, labels ...string) prmmet.Metric {
	m := prmmet.NewMetrics(name, prmtps.Counter)
	m.SetDesc("Test counter metric")
	if len(labels) > 0 {
		m.AddLabel(labels...)
	}
	// Set a minimal collect function to satisfy Add requirement
	m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {})
	return m
}

// Helper function to create a gauge metric
func createGaugeMetric(name string, labels ...string) prmmet.Metric {
	m := prmmet.NewMetrics(name, prmtps.Gauge)
	m.SetDesc("Test gauge metric")
	if len(labels) > 0 {
		m.AddLabel(labels...)
	}
	m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {})
	return m
}

// Helper function to create a histogram metric
func createHistogramMetric(name string, labels ...string) prmmet.Metric {
	m := prmmet.NewMetrics(name, prmtps.Histogram)
	m.SetDesc("Test histogram metric")
	m.AddBuckets(prmsdk.DefBuckets...)
	if len(labels) > 0 {
		m.AddLabel(labels...)
	}
	m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {})
	return m
}

// Helper function to create a summary metric
func createSummaryMetric(name string, labels ...string) prmmet.Metric {
	m := prmmet.NewMetrics(name, prmtps.Summary)
	m.SetDesc("Test summary metric")
	m.AddObjective(0.5, 0.05)
	m.AddObjective(0.9, 0.01)
	m.AddObjective(0.99, 0.001)
	if len(labels) > 0 {
		m.AddLabel(labels...)
	}
	m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {})
	return m
}

// Helper function to create and add a metric to the pool
func addMetricToPool(pool prmpool.MetricPool, name string, metricType prmtps.MetricType, labels ...string) prmmet.Metric {
	var m prmmet.Metric
	switch metricType {
	case prmtps.Counter:
		m = createCounterMetric(name, labels...)
	case prmtps.Gauge:
		m = createGaugeMetric(name, labels...)
	case prmtps.Histogram:
		m = createHistogramMetric(name, labels...)
	case prmtps.Summary:
		m = createSummaryMetric(name, labels...)
	default:
		return nil
	}

	err := pool.Add(m)
	Expect(err).ToNot(HaveOccurred())
	return m
}

// Helper function to generate unique metric name
func uniqueMetricName(base string) string {
	count := metricCounter.Add(1)
	return fmt.Sprintf("%s_%d", base, count)
}

// Helper function to cleanup metrics from registry
func cleanupMetric(m prmmet.Metric) {
	if m != nil {
		m.UnRegister(testRegistry)
	}
}
