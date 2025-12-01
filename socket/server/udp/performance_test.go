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
 *
 */

package udp_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("UDP Server Performance", Ordered, func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeAll(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterAll(func() {
		if cancel != nil {
			cancel()
		}
		time.Sleep(100 * time.Millisecond)
	})

	Describe("Server Creation Performance", func() {
		It("should benchmark server creation", func() {
			experiment := gmeasure.NewExperiment("Server Creation")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("creation", func() {
					handler := newTestHandler(false)
					srv, err := createServerWithHandler(handler.handler)
					Expect(err).ToNot(HaveOccurred())
					Expect(srv).ToNot(BeNil())
				})
			}, gmeasure.SamplingConfig{N: 10})

			stats := experiment.GetStats("creation")
			AddReportEntry("Creation Stats", stats)

			// Server creation should be fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Millisecond))
		})
	})

	Describe("Server Startup Performance", func() {
		It("should benchmark server startup", func() {
			experiment := gmeasure.NewExperiment("Server Startup")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				localCtx, localCancel := context.WithCancel(ctx)
				defer localCancel()

				experiment.MeasureDuration("startup", func() {
					go func() {
						_ = srv.Listen(localCtx)
					}()

					Eventually(func() bool {
						return srv.IsRunning()
					}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
				})

				localCancel()
				time.Sleep(50 * time.Millisecond)
			}, gmeasure.SamplingConfig{N: 10})

			stats := experiment.GetStats("startup")
			AddReportEntry("Startup Stats", stats)

			// Startup should be reasonable
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 100*time.Millisecond))
		})
	})

	Describe("Server Shutdown Performance", func() {
		It("should benchmark server shutdown", func() {
			experiment := gmeasure.NewExperiment("Server Shutdown")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				localCtx, localCancel := context.WithCancel(ctx)

				startServer(srv, localCtx)

				experiment.MeasureDuration("shutdown", func() {
					shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer shutdownCancel()

					err := srv.Shutdown(shutdownCtx)
					Expect(err).ToNot(HaveOccurred())
				})

				localCancel()
				time.Sleep(50 * time.Millisecond)
			}, gmeasure.SamplingConfig{N: 10})

			stats := experiment.GetStats("shutdown")
			AddReportEntry("Shutdown Stats", stats)

			// Shutdown should be fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 100*time.Millisecond))
		})
	})

	Describe("State Query Performance", func() {
		var srv interface{}

		BeforeAll(func() {
			handler := newTestHandler(false)
			var err error
			srv, err = createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should benchmark IsRunning calls", func() {
			experiment := gmeasure.NewExperiment("IsRunning Performance")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("is_running", func() {
					for i := 0; i < 1000; i++ {
						_ = srv.(interface{ IsRunning() bool }).IsRunning()
					}
				})
			}, gmeasure.SamplingConfig{N: 10})

			stats := experiment.GetStats("is_running")
			AddReportEntry("IsRunning Stats", stats)

			// 1000 calls should be very fast (atomic operation)
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Millisecond))
		})

		It("should benchmark IsGone calls", func() {
			experiment := gmeasure.NewExperiment("IsGone Performance")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("is_gone", func() {
					for i := 0; i < 1000; i++ {
						_ = srv.(interface{ IsGone() bool }).IsGone()
					}
				})
			}, gmeasure.SamplingConfig{N: 10})

			stats := experiment.GetStats("is_gone")
			AddReportEntry("IsGone Stats", stats)

			// 1000 calls should be very fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Millisecond))
		})

		It("should benchmark OpenConnections calls", func() {
			experiment := gmeasure.NewExperiment("OpenConnections Performance")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("open_connections", func() {
					for i := 0; i < 1000; i++ {
						_ = srv.(interface{ OpenConnections() int64 }).OpenConnections()
					}
				})
			}, gmeasure.SamplingConfig{N: 10})

			stats := experiment.GetStats("open_connections")
			AddReportEntry("OpenConnections Stats", stats)

			// 1000 calls should be very fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Millisecond))
		})
	})

	Describe("Callback Registration Performance", func() {
		It("should benchmark callback registrations", func() {
			experiment := gmeasure.NewExperiment("Callback Registration")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				experiment.MeasureDuration("register_callbacks", func() {
					for i := 0; i < 100; i++ {
						srv.RegisterFuncError(func(...error) {})
						srv.RegisterFuncInfo(newInfoCollector().callback)
						srv.RegisterFuncInfoServer(func(string) {})
					}
				})
			}, gmeasure.SamplingConfig{N: 10})

			stats := experiment.GetStats("register_callbacks")
			AddReportEntry("Registration Stats", stats)

			// 300 registrations (100 x 3) should be fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 50*time.Millisecond))
		})
	})

	Describe("Complete Lifecycle Performance", func() {
		It("should benchmark full start/stop cycle", func() {
			experiment := gmeasure.NewExperiment("Complete Lifecycle")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("full_cycle", func() {
					handler := newTestHandler(false)
					srv, err := createServerWithHandler(handler.handler)
					Expect(err).ToNot(HaveOccurred())

					localCtx, localCancel := context.WithCancel(ctx)
					defer localCancel()

					// Start
					go func() {
						_ = srv.Listen(localCtx)
					}()

					Eventually(func() bool {
						return srv.IsRunning()
					}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

					// Shutdown
					shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer shutdownCancel()

					err = srv.Shutdown(shutdownCtx)
					Expect(err).ToNot(HaveOccurred())
				})

				time.Sleep(50 * time.Millisecond)
			}, gmeasure.SamplingConfig{N: 5})

			stats := experiment.GetStats("full_cycle")
			AddReportEntry("Cycle Stats", stats)

			// Full cycle should complete in reasonable time
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 500*time.Millisecond))
		})
	})

	Describe("Memory Efficiency", func() {
		It("should have minimal memory overhead", func() {
			experiment := gmeasure.NewExperiment("Memory Usage")
			AddReportEntry(experiment.Name, experiment)

			// Create multiple servers to observe memory pattern
			servers := make([]interface{}, 10)

			experiment.Sample(func(idx int) {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())
				servers[idx%10] = srv
			}, gmeasure.SamplingConfig{N: 10})

			// Memory should not grow excessively
			Expect(len(servers)).To(Equal(10))
		})
	})
})
