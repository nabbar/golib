//go:build linux || darwin

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

package unixgram_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Unix Datagram Server Performance", Ordered, func() {
	var (
		ctx      context.Context
		cancel   context.CancelFunc
		sockPath string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
		cleanupSocketFile(sockPath)
		time.Sleep(50 * time.Millisecond)
	})

	Describe("Server Startup Performance", func() {
		var experiment *Experiment

		BeforeAll(func() {
			experiment = NewExperiment("Server Startup")
			AddReportEntry(experiment.Name, experiment)
		})

		It("should measure startup time", func() {
			experiment.Sample(func(idx int) {
				handler := newTestHandler(false)
				srv, path, err := createServerWithHandler(handler.handler)
				sockPath = path
				Expect(err).ToNot(HaveOccurred())
				defer srv.Close()

				experiment.MeasureDuration("start", func() {
					startServer(srv, ctx)
				})

				cancel()
				ctx, cancel = context.WithCancel(testCtx)
			}, SamplingConfig{N: 20, Duration: 10 * time.Second})

			stats := experiment.GetStats("start")
			AddReportEntry("Startup Stats", stats)

			Expect(stats.DurationFor(StatMedian)).To(BeNumerically("<", 50*time.Millisecond),
				"Median startup time should be under 50ms")
		})
	})

	Describe("Server Shutdown Performance", func() {
		var experiment *Experiment

		BeforeAll(func() {
			experiment = NewExperiment("Server Shutdown")
			AddReportEntry(experiment.Name, experiment)
		})

		It("should measure shutdown time", func() {
			experiment.Sample(func(idx int) {
				handler := newTestHandler(false)
				srv, path, err := createServerWithHandler(handler.handler)
				sockPath = path
				Expect(err).ToNot(HaveOccurred())

				startServer(srv, ctx)

				experiment.MeasureDuration("shutdown", func() {
					shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer shutdownCancel()
					_ = srv.Shutdown(shutdownCtx)
				})

				cleanupSocketFile(sockPath)
				cancel()
				ctx, cancel = context.WithCancel(testCtx)
			}, SamplingConfig{N: 20, Duration: 10 * time.Second})

			stats := experiment.GetStats("shutdown")
			AddReportEntry("Shutdown Stats", stats)

			Expect(stats.DurationFor(StatMedian)).To(BeNumerically("<", 100*time.Millisecond),
				"Median shutdown time should be under 100ms")
		})
	})

	Describe("Datagram Throughput", func() {
		var experiment *Experiment

		BeforeAll(func() {
			experiment = NewExperiment("Datagram Throughput")
			AddReportEntry(experiment.Name, experiment)
		})

		It("should measure datagram processing throughput", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			experiment.Sample(func(idx int) {
				numDatagrams := 100
				data := []byte("test message")

				experiment.MeasureDuration("send_100", func() {
					for i := 0; i < numDatagrams; i++ {
						_ = sendUnixgramDatagram(sockPath, data)
					}
				})

				// Wait for all to be received
				Eventually(func() int64 {
					return handler.getCount()
				}, 5*time.Second, 10*time.Millisecond).Should(BeNumerically(">=", int64(numDatagrams)))
			}, SamplingConfig{N: 10, Duration: 10 * time.Second})

			stats := experiment.GetStats("send_100")
			AddReportEntry("Throughput Stats", stats)

			// 100 datagrams should be sent in reasonable time
			Expect(stats.DurationFor(StatMedian)).To(BeNumerically("<", 1*time.Second),
				"Sending 100 datagrams should take less than 1 second")
		})
	})

	Describe("State Query Performance", func() {
		var experiment *Experiment

		BeforeAll(func() {
			experiment = NewExperiment("State Query Performance")
			AddReportEntry(experiment.Name, experiment)
		})

		It("should measure state query latency", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("is_running", func() {
					_ = srv.IsRunning()
				})

				experiment.MeasureDuration("is_gone", func() {
					_ = srv.IsGone()
				})

				experiment.MeasureDuration("open_connections", func() {
					_ = srv.OpenConnections()
				})
			}, SamplingConfig{N: 1000, Duration: 5 * time.Second})

			for _, metric := range []string{"is_running", "is_gone", "open_connections"} {
				stats := experiment.GetStats(metric)
				AddReportEntry(metric+" Stats", stats)

				// State queries should be very fast (< 1µs median)
				Expect(stats.DurationFor(StatMedian)).To(BeNumerically("<", 10*time.Microsecond),
					metric+" should be very fast")
			}
		})
	})

	Describe("Large Datagram Performance", func() {
		var experiment *Experiment

		BeforeAll(func() {
			experiment = NewExperiment("Large Datagram Performance")
			AddReportEntry(experiment.Name, experiment)
		})

		It("should measure large datagram handling", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			largeData := make([]byte, 16384) // 16KB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("send_16kb", func() {
					_ = sendUnixgramDatagram(sockPath, largeData)
				})

				// Wait for datagram to be received
				Eventually(func() int64 {
					return handler.getCount()
				}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">", 0))
			}, SamplingConfig{N: 20, Duration: 10 * time.Second})

			stats := experiment.GetStats("send_16kb")
			AddReportEntry("Large Datagram Stats", stats)

			Expect(stats.DurationFor(StatMedian)).To(BeNumerically("<", 10*time.Millisecond),
				"Sending 16KB datagram should take less than 10ms")
		})
	})

	Describe("Callback Overhead", func() {
		var experiment *Experiment

		BeforeAll(func() {
			experiment = NewExperiment("Callback Overhead")
			AddReportEntry(experiment.Name, experiment)
		})

		It("should measure callback registration overhead", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			experiment.Sample(func(idx int) {
				collector := newInfoCollector()

				experiment.MeasureDuration("register_callback", func() {
					srv.RegisterFuncInfo(collector.callback)
				})
			}, SamplingConfig{N: 100, Duration: 5 * time.Second})

			stats := experiment.GetStats("register_callback")
			AddReportEntry("Callback Registration Stats", stats)

			Expect(stats.DurationFor(StatMedian)).To(BeNumerically("<", 1*time.Microsecond),
				"Callback registration should be very fast")
		})
	})
})
