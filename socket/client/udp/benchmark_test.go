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
	"net"
	"time"

	libsck "github.com/nabbar/golib/socket"
	"github.com/onsi/gomega/gmeasure"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Client Performance", Ordered, func() {
	var experiment *gmeasure.Experiment

	BeforeAll(func() {
		experiment = gmeasure.NewExperiment("UDP Client Performance")
		AddReportEntry(experiment.Name, experiment)
	})

	Describe("Client Creation Performance", func() {
		It("should measure client creation time", func() {
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("client_creation", func() {
					cli := createClient(getTestAddress())
					_ = cli
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 3 * time.Second})

			stats := experiment.GetStats("client_creation")
			AddReportEntry("Creation Stats", stats)

			// Should be very fast (microseconds)
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 1*time.Millisecond))
		})
	})

	Describe("Connection Performance", func() {
		It("should measure connect time", func() {
			srv, _, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			address := getTestAddress()
			srv2 := createServer(simpleEchoHandler(), address)
			startServer(ctx, srv2)
			waitForServerRunning(address, 5*time.Second)
			defer cleanupServer(srv2, ctx)

			experiment.Sample(func(idx int) {
				cli := createClient(address)
				defer cleanupClient(cli)

				experiment.MeasureDuration("connect", func() {
					err := cli.Connect(ctx)
					Expect(err).ToNot(HaveOccurred())
				})
			}, gmeasure.SamplingConfig{N: 50, Duration: 5 * time.Second})

			stats := experiment.GetStats("connect")
			AddReportEntry("Connect Stats", stats)

			// Connect should be relatively fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Millisecond))
		})
	})

	Describe("Write Performance", func() {
		// Note: UDP writes may fail with "connection refused" if server isn't ready
		// This is a timing issue in benchmarks and doesn't affect real-world usage
		It("should measure small datagram write time", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			data := []byte("small message")

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("write_small", func() {
					cli.Write(data) // Ignore errors - UDP may lose packets
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 3 * time.Second})

			stats := experiment.GetStats("write_small")
			AddReportEntry("Small Write Stats", stats)

			// Writes should be fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 5*time.Millisecond))
		})

		It("should measure large datagram write time", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			data := make([]byte, 1400) // Safe MTU size

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("write_large", func() {
					cli.Write(data) // Ignore errors - UDP may lose packets
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 3 * time.Second})

			stats := experiment.GetStats("write_large")
			AddReportEntry("Large Write Stats", stats)

			// Large writes might be slightly slower but still fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Millisecond))
		})
	})

	Describe("Throughput Performance", func() {
		It("should measure sequential write throughput", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			data := []byte("throughput test message")
			messagesPerSample := 100

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("throughput_100msgs", func() {
					for i := 0; i < messagesPerSample; i++ {
						_, _ = cli.Write(data)
					}
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 5 * time.Second})

			stats := experiment.GetStats("throughput_100msgs")
			AddReportEntry("Throughput Stats", stats)

			// 100 messages should complete reasonably fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 500*time.Millisecond))
		})
	})

	Describe("State Check Performance", func() {
		It("should measure IsConnected performance", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("is_connected", func() {
					_ = cli.IsConnected()
				})
			}, gmeasure.SamplingConfig{N: 1000, Duration: 3 * time.Second})

			stats := experiment.GetStats("is_connected")
			AddReportEntry("IsConnected Stats", stats)

			// State checks should be extremely fast (nanoseconds)
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 100*time.Microsecond))
		})
	})

	Describe("Close Performance", func() {
		It("should measure close time", func() {
			srv, _, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			address := getTestAddress()
			srv2 := createServer(simpleEchoHandler(), address)
			startServer(ctx, srv2)
			waitForServerRunning(address, 5*time.Second)
			defer cleanupServer(srv2, ctx)

			experiment.Sample(func(idx int) {
				cli := createClient(address)
				_ = cli.Connect(ctx)

				experiment.MeasureDuration("close", func() {
					err := cli.Close()
					Expect(err).ToNot(HaveOccurred())
				})
			}, gmeasure.SamplingConfig{N: 50, Duration: 5 * time.Second})

			stats := experiment.GetStats("close")
			AddReportEntry("Close Stats", stats)

			// Close should be fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 5*time.Millisecond))
		})
	})

	Describe("Complete Cycle Performance", func() {
		It("should measure full create-connect-write-close cycle", func() {
			srv, _, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			address := getTestAddress()
			srv2 := createServer(simpleEchoHandler(), address)
			startServer(ctx, srv2)
			waitForServerRunning(address, 5*time.Second)
			defer cleanupServer(srv2, ctx)

			data := []byte("cycle test")

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("full_cycle", func() {
					cli := createClient(address)
					_ = cli.Connect(ctx)
					_, _ = cli.Write(data)
					_ = cli.Close()
				})
			}, gmeasure.SamplingConfig{N: 50, Duration: 5 * time.Second})

			stats := experiment.GetStats("full_cycle")
			AddReportEntry("Full Cycle Stats", stats)

			// Complete cycle should complete reasonably fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 20*time.Millisecond))
		})
	})

	Describe("Callback Registration Performance", func() {
		It("should measure callback registration time", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("register_callback", func() {
					cli.RegisterFuncError(func(_ ...error) {})
					cli.RegisterFuncInfo(func(_, _ net.Addr, _ libsck.ConnState) {})
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 3 * time.Second})

			stats := experiment.GetStats("register_callback")
			AddReportEntry("Callback Registration Stats", stats)

			// Registration should be very fast
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 1*time.Millisecond))
		})
	})
})
