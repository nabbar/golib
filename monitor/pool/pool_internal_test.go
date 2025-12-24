/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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
 */

package pool_test

import (
	"bytes"
	"context"
	"io"
	"time"

	monpool "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"
	libprm "github.com/nabbar/golib/prometheus"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool Internal", func() {
	var (
		pool monpool.Pool
		ctx  context.Context
		cnl  context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(context.Background(), 10*time.Second)
		pool = monpool.New(ctx)
	})

	AfterEach(func() {
		if cnl != nil {
			cnl()
		}
	})

	Describe("Internal Metrics Collection", func() {
		It("should handle metric observation errors gracefully", func() {
			mockMet := &monpool.MockMetric{}
			mon := &monpool.MockMonitor{NameVal: "test-mon"}

			// Inject monitor
			monpool.PoolInjectStore(pool, mon.Name(), mon)

			// Execute collect functions - should not panic despite mock returning errors
			monpool.PoolCollectMetricLatency(pool, ctx, mockMet)
			monpool.PoolCollectMetricUptime(pool, ctx, mockMet)
			monpool.PoolCollectMetricDowntime(pool, ctx, mockMet)
			monpool.PoolCollectMetricRiseTime(pool, ctx, mockMet)
			monpool.PoolCollectMetricFallTime(pool, ctx, mockMet)
			monpool.PoolCollectMetricStatus(pool, ctx, mockMet)
			monpool.PoolCollectMetricRising(pool, ctx, mockMet)
			monpool.PoolCollectMetricFalling(pool, ctx, mockMet)
			monpool.PoolCollectMetricSLis(pool, ctx, mockMet)
		})
	})

	Describe("Internal Metrics Creation", func() {
		It("should return error when Prometheus AddMetric fails", func() {
			mockProm := &monpool.MockPrometheus{}

			monpool.PoolSetPrometheusFactory(pool, func() libprm.Prometheus {
				return mockProm
			})

			err := monpool.PoolCreateMetrics(pool)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mock add metric error"))
		})
	})

	Describe("Internal Storage Integrity", func() {
		It("should remove invalid types during MonitorWalk", func() {
			// Inject invalid type
			monpool.PoolInjectStore(pool, "invalid", "not-a-monitor")

			// Verify it exists
			val, ok := monpool.PoolLoadStore(pool, "invalid")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("not-a-monitor"))

			// Run Walk which should clean up invalid types
			// The function is not called for invalid types, but the internal walk deletes them
			pool.MonitorWalk(func(_ string, _ montps.Monitor) bool { return true })

			// Verify it was deleted
			_, ok = monpool.PoolLoadStore(pool, "invalid")
			Expect(ok).To(BeFalse(), "Invalid item should be removed from store")
		})
	})

	Describe("Internal Lifecycle Errors", func() {
		It("should report errors when MonitorSet fails during Start/Stop/Restart", func() {
			// Monitor with empty name will cause MonitorSet to fail
			mon := &monpool.MockMonitor{NameVal: ""}

			// Store with a key so Walk finds it
			monpool.PoolInjectStore(pool, "key-for-empty", mon)

			// Start
			err := pool.Start(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error on starting monitor"))

			// Stop
			err = pool.Stop(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error on stopping monitor"))

			// Restart
			err = pool.Restart(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error on restarting monitor"))
		})
	})

	Describe("Internal Shell Commands", func() {
		var (
			mon *monpool.MockMonitor
		)

		BeforeEach(func() {
			mon = &monpool.MockMonitor{NameVal: ""}
			monpool.PoolInjectStore(pool, "key-for-empty", mon)
		})

		It("should output errors to stderr during Start command execution", func() {
			cmd := monpool.PoolShlCmdStart(pool, ctx)
			Expect(cmd).ToNot(BeNil())

			var errBuf bytes.Buffer
			cmd.Run(io.Discard, &errBuf, []string{"key-for-empty"})
			Expect(errBuf.String()).ToNot(BeEmpty())
		})

		It("should output errors to stderr during Stop command execution", func() {
			cmd := monpool.PoolShlCmdStop(pool, ctx)
			Expect(cmd).ToNot(BeNil())

			var errBuf bytes.Buffer
			cmd.Run(io.Discard, &errBuf, []string{"key-for-empty"})
			Expect(errBuf.String()).ToNot(BeEmpty())
		})

		It("should output errors to stderr during Restart command execution", func() {
			cmd := monpool.PoolShlCmdRestart(pool, ctx)
			Expect(cmd).ToNot(BeNil())

			var errBuf bytes.Buffer
			cmd.Run(io.Discard, &errBuf, []string{"key-for-empty"})
			Expect(errBuf.String()).ToNot(BeEmpty())
		})
	})
})
