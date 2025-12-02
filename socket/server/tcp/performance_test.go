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

// performance_test.go measures server performance characteristics using gmeasure.
// Benchmarks include server startup/shutdown time, connection establishment latency,
// echo latency, throughput, and concurrent connection handling capacity.
package tcp_test

import (
	"context"
	"time"

	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gmeasure"
)

var _ = Describe("TCP Server Performance", Ordered, func() {
	var (
		exp *Experiment
		srv scksrt.ServerTcp
		adr string
		c   context.Context
		cnl context.CancelFunc
	)

	BeforeAll(func() {
		exp = NewExperiment("TCP Server Performance")
		AddReportEntry(exp.Name, exp)
	})

	BeforeEach(func() {
		adr = getTestAddr()
		c, cnl = context.WithCancel(globalCtx)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		time.Sleep(100 * time.Millisecond)
	})

	Context("server lifecycle", func() {
		It("should measure server startup time", func() {
			exp.Sample(func(idx int) {
				// Use a unique address for each sample to avoid port conflicts
				testAddr := getTestAddr()
				cfg := createDefaultConfig(testAddr)
				var err error
				var testSrv scksrt.ServerTcp

				exp.MeasureDuration("startup", func() {
					testSrv, err = scksrt.New(nil, echoHandler, cfg)
					Expect(err).ToNot(HaveOccurred())
					startServerInBackground(c, testSrv)
					waitForServer(testSrv, 5*time.Second)
				})

				// Cleanup this test server
				if testSrv != nil {
					_ = testSrv.Close()
					time.Sleep(10 * time.Millisecond)
				}
			}, SamplingConfig{N: 10})

			stts := exp.GetStats("startup")
			AddReportEntry("Startup Stats", stts)
			Expect(stts.DurationFor(StatMedian)).To(BeNumerically("<", time.Second))
		})

		It("should measure server shutdown time", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			exp.Sample(func(idx int) {
				exp.MeasureDuration("shutdown", func() {
					err := srv.Shutdown(c)
					Expect(err).ToNot(HaveOccurred())
				})
			}, SamplingConfig{N: 1})

			stts := exp.GetStats("shutdown")
			AddReportEntry("Shutdown Stats", stts)
			Expect(stts.DurationFor(StatMedian)).To(BeNumerically("<", 2*time.Second))
		})
	})

	Context("connection handling", func() {
		BeforeEach(func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)
		})

		It("should measure connection establishment time", func() {
			exp.Sample(func(idx int) {
				exp.MeasureDuration("connect", func() {
					con := connectToServer(adr)
					_ = con.Close()
				})
			}, SamplingConfig{N: 20})

			stts := exp.GetStats("connect")
			AddReportEntry("Connection Stats", stts)
			Expect(stts.DurationFor(StatMedian)).To(BeNumerically("<", 100*time.Millisecond))
		})

		It("should measure echo latency", func() {
			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			exp.Sample(func(idx int) {
				msg := []byte("test message for latency measurement")

				exp.MeasureDuration("echo", func() {
					_ = sendAndReceive(con, msg)
				})
			}, SamplingConfig{N: 50})

			stts := exp.GetStats("echo")
			AddReportEntry("Echo Latency Stats", stts)
			Expect(stts.DurationFor(StatMedian)).To(BeNumerically("<", 50*time.Millisecond))
		})

		It("should measure throughput", func() {
			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			msgCount := 100
			msg := []byte("throughput test message")

			exp.Sample(func(idx int) {
				exp.MeasureDuration("throughput", func() {
					for i := 0; i < msgCount; i++ {
						_ = sendAndReceive(con, msg)
					}
				})
			}, SamplingConfig{N: 5})

			stts := exp.GetStats("throughput")
			AddReportEntry("Throughput Stats", stts)
			Expect(stts.DurationFor(StatMedian)).To(BeNumerically("<", 5*time.Second))
		})
	})

	Context("concurrent operations", func() {
		BeforeEach(func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)
		})

		It("should measure concurrent connection handling", func() {
			numConns := 10

			exp.Sample(func(idx int) {
				exp.MeasureDuration("concurrent", func() {
					var conns []any
					for i := 0; i < numConns; i++ {
						con := connectToServer(adr)
						conns = append(conns, con)

						msg := []byte("concurrent test")
						_ = sendAndReceive(con, msg)
					}

					for _, c := range conns {
						if closer, ok := c.(interface{ Close() error }); ok {
							_ = closer.Close()
						}
					}
				})
			}, SamplingConfig{N: 10})

			stts := exp.GetStats("concurrent")
			AddReportEntry("Concurrent Stats", stts)
			Expect(stts.DurationFor(StatMedian)).To(BeNumerically("<", 2*time.Second))
		})
	})
})
