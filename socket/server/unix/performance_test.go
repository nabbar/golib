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

// performance_test.go provides performance benchmarks for the Unix socket server.
// Measures throughput, latency, and resource usage using gmeasure.
package unix_test

import (
	"context"
	"sync"
	"time"

	scksru "github.com/nabbar/golib/socket/server/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Unix Server Performance Benchmarks", Ordered, func() {
	var (
		srv        scksru.ServerUnix
		socketPath string
		c          context.Context
		cnl        context.CancelFunc
		experiment *gmeasure.Experiment
	)

	BeforeAll(func() {
		experiment = gmeasure.NewExperiment("Unix Server Performance")
		AddReportEntry(experiment.Name, experiment)
	})

	BeforeEach(func() {
		socketPath = getTestSocketPath()
		c, cnl = context.WithCancel(globalCtx)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		cleanupSocketFile(socketPath)
		time.Sleep(100 * time.Millisecond)
	})

	Context("server startup and shutdown", func() {
		It("should measure server start time", func() {
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("start", func() {
					cfg := createDefaultConfig(socketPath)
					var err error
					srv, err = scksru.New(nil, echoHandler, cfg)
					Expect(err).ToNot(HaveOccurred())

					startServerInBackground(c, srv)
					waitForServer(srv, 2*time.Second)
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 30 * time.Second})

			stats := experiment.GetStats("start")
			AddReportEntry("Start Time Stats", stats)

			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 100*time.Millisecond))
		})

		It("should measure server stop time", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("stop", func() {
					err := srv.Shutdown(c)
					Expect(err).ToNot(HaveOccurred())
				})

				// Restart for next sample
				if idx < 9 {
					startServerInBackground(c, srv)
					waitForServer(srv, 2*time.Second)
				}
			}, gmeasure.SamplingConfig{N: 10, Duration: 30 * time.Second})

			stats := experiment.GetStats("stop")
			AddReportEntry("Stop Time Stats", stats)

			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 200*time.Millisecond))
		})
	})

	Context("connection throughput", func() {
		It("should measure single connection throughput", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			experiment.Sample(func(idx int) {
				con := connectToServer(socketPath)
				defer func() { _ = con.Close() }()

				experiment.MeasureDuration("single-conn-echo", func() {
					for i := 0; i < 100; i++ {
						data := []byte("Test message")
						_ = sendAndReceive(con, data)
					}
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 30 * time.Second})

			stats := experiment.GetStats("single-conn-echo")
			AddReportEntry("Single Connection Throughput", stats)
		})

		It("should measure concurrent connection throughput", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("concurrent-10-conns", func() {
					wg := sync.WaitGroup{}
					wg.Add(10)

					for i := 0; i < 10; i++ {
						go func() {
							defer wg.Done()

							con := connectToServer(socketPath)
							defer func() { _ = con.Close() }()

							for j := 0; j < 10; j++ {
								data := []byte("Concurrent test")
								_ = sendAndReceive(con, data)
							}
						}()
					}

					wg.Wait()
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 30 * time.Second})

			stats := experiment.GetStats("concurrent-10-conns")
			AddReportEntry("Concurrent Connections (10)", stats)
		})
	})

	Context("connection latency", func() {
		It("should measure connection establishment latency", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("conn-establish", func() {
					con := connectToServer(socketPath)
					_ = con.Close()
				})
			}, gmeasure.SamplingConfig{N: 50, Duration: 30 * time.Second})

			stats := experiment.GetStats("conn-establish")
			AddReportEntry("Connection Establishment Latency", stats)

			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Millisecond))
		})

		It("should measure echo round-trip latency", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			experiment.Sample(func(idx int) {
				data := []byte("Latency test message")

				experiment.MeasureDuration("echo-rtt", func() {
					_ = sendAndReceive(con, data)
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 30 * time.Second})

			stats := experiment.GetStats("echo-rtt")
			AddReportEntry("Echo Round-Trip Time", stats)

			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 5*time.Millisecond))
		})
	})

	Context("scalability", func() {
		It("should measure performance with many connections", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			experiment.Sample(func(idx int) {
				const numConns = 50

				experiment.MeasureDuration("many-conns-50", func() {
					wg := sync.WaitGroup{}
					wg.Add(numConns)

					for i := 0; i < numConns; i++ {
						go func() {
							defer wg.Done()

							con := connectToServer(socketPath)
							defer func() { _ = con.Close() }()

							data := []byte("Load test")
							_ = sendAndReceive(con, data)
						}()
					}

					wg.Wait()
				})
			}, gmeasure.SamplingConfig{N: 5, Duration: 30 * time.Second})

			stats := experiment.GetStats("many-conns-50")
			AddReportEntry("Scalability (50 connections)", stats)
		})
	})

	Context("message size impact", func() {
		It("should measure small message performance", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			smallData := []byte("Small")

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("echo-small", func() {
					for i := 0; i < 100; i++ {
						_ = sendAndReceive(con, smallData)
					}
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 30 * time.Second})

			stats := experiment.GetStats("echo-small")
			AddReportEntry("Small Messages (5 bytes)", stats)
		})

		It("should measure large message performance", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			largeData := make([]byte, 8192)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("echo-large", func() {
					for i := 0; i < 10; i++ {
						_ = sendAndReceive(con, largeData)
					}
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 30 * time.Second})

			stats := experiment.GetStats("echo-large")
			AddReportEntry("Large Messages (8KB)", stats)
		})
	})
})
