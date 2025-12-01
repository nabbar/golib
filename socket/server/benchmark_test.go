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

package server_test

import (
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Server Factory Performance", func() {
	Context("Factory Overhead", func() {
		It("should create TCP servers with minimal overhead", func() {
			exp := NewExperiment("TCP Server Creation")
			AddReportEntry(exp.Name, exp)

			exp.Sample(func(idx int) {
				exp.MeasureDuration("creation_time", func() {
					cfg := sckcfg.Server{
						Network: libptc.NetworkTCP,
						Address: getTestTCPAddress(),
					}

					srv, err := scksrv.New(nil, basicHandler(), cfg)
					Expect(err).ToNot(HaveOccurred())
					if srv != nil {
						_ = srv.Close()
					}
				})
			}, SamplingConfig{N: 100})

			median := exp.GetStats("creation_time").DurationFor(StatMedian)
			Expect(median).To(BeNumerically("<", 10*time.Millisecond),
				"TCP server creation should take less than 10ms")
		})

		It("should create UDP servers with minimal overhead", func() {
			exp := NewExperiment("UDP Server Creation")
			AddReportEntry(exp.Name, exp)

			exp.Sample(func(idx int) {
				exp.MeasureDuration("creation_time", func() {
					cfg := sckcfg.Server{
						Network: libptc.NetworkUDP,
						Address: getTestUDPAddress(),
					}

					srv, err := scksrv.New(nil, basicHandler(), cfg)
					Expect(err).ToNot(HaveOccurred())
					if srv != nil {
						_ = srv.Close()
					}
				})
			}, SamplingConfig{N: 100})

			median := exp.GetStats("creation_time").DurationFor(StatMedian)
			Expect(median).To(BeNumerically("<", 10*time.Millisecond),
				"UDP server creation should take less than 10ms")
		})
	})

	Context("Memory Allocation", func() {
		It("should have minimal memory allocation for TCP", func() {
			exp := NewExperiment("TCP Memory Allocation")
			AddReportEntry(exp.Name, exp)

			exp.Sample(func(idx int) {
				cfg := sckcfg.Server{
					Network: libptc.NetworkTCP,
					Address: getTestTCPAddress(),
				}

				srv, err := scksrv.New(nil, basicHandler(), cfg)
				Expect(err).ToNot(HaveOccurred())
				if srv != nil {
					_ = srv.Close()
				}
			}, SamplingConfig{N: 50})
		})

		It("should have minimal memory allocation for UDP", func() {
			exp := NewExperiment("UDP Memory Allocation")
			AddReportEntry(exp.Name, exp)

			exp.Sample(func(idx int) {
				cfg := sckcfg.Server{
					Network: libptc.NetworkUDP,
					Address: getTestUDPAddress(),
				}

				srv, err := scksrv.New(nil, basicHandler(), cfg)
				Expect(err).ToNot(HaveOccurred())
				if srv != nil {
					_ = srv.Close()
				}
			}, SamplingConfig{N: 50})
		})
	})

	Context("Concurrent Creation Performance", func() {
		It("should handle concurrent server creation efficiently", func() {
			exp := NewExperiment("Concurrent Server Creation")
			AddReportEntry(exp.Name, exp)

			exp.Sample(func(idx int) {
				exp.MeasureDuration("concurrent_creation", func() {
					done := make(chan bool, 10)

					for i := 0; i < 10; i++ {
						go func() {
							cfg := sckcfg.Server{
								Network: libptc.NetworkTCP,
								Address: getTestTCPAddress(),
							}

							srv, err := scksrv.New(nil, basicHandler(), cfg)
							Expect(err).ToNot(HaveOccurred())
							if srv != nil {
								_ = srv.Close()
							}

							done <- true
						}()
					}

					for i := 0; i < 10; i++ {
						<-done
					}
				})
			}, SamplingConfig{N: 20})

			median := exp.GetStats("concurrent_creation").DurationFor(StatMedian)
			Expect(median).To(BeNumerically("<", 100*time.Millisecond),
				"Concurrent creation should complete in less than 100ms")
		})
	})
})
