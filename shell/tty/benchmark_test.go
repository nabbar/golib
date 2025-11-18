/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package tty_test

import (
	"time"

	"github.com/nabbar/golib/shell/tty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Benchmarks", func() {
	Describe("Restore Performance", func() {
		Context("single restore operations", func() {
			It("should measure Restore with nil", func() {
				experiment := gmeasure.NewExperiment("Restore - Nil State")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("restore-nil", func() {
						tty.Restore(nil)
					})
				}, gmeasure.SamplingConfig{N: 1000, Duration: 5 * time.Second})

				Expect(experiment.Get("restore-nil").Stats().DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 1*time.Microsecond))
			})

			It("should measure Restore with valid mock", func() {
				experiment := gmeasure.NewExperiment("Restore - Valid Mock")
				AddReportEntry(experiment.Name, experiment)

				mock := newMockTTYSaver(false)

				experiment.Sample(func(idx int) {
					mock.Reset()
					experiment.MeasureDuration("restore-valid", func() {
						tty.Restore(mock)
					})
				}, gmeasure.SamplingConfig{N: 1000, Duration: 5 * time.Second})

				Expect(experiment.Get("restore-valid").Stats().DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Microsecond))
			})

			It("should measure Restore with failing mock", func() {
				experiment := gmeasure.NewExperiment("Restore - Failing Mock")
				AddReportEntry(experiment.Name, experiment)

				mock := newMockTTYSaver(true)

				experiment.Sample(func(idx int) {
					mock.Reset()
					experiment.MeasureDuration("restore-fail", func() {
						tty.Restore(mock)
					})
				}, gmeasure.SamplingConfig{N: 1000, Duration: 5 * time.Second})

				// Failing restore should still be fast
				Expect(experiment.Get("restore-fail").Stats().DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Microsecond))
			})
		})

		Context("concurrent restore operations", func() {
			It("should measure concurrent Restore calls", func() {
				experiment := gmeasure.NewExperiment("Restore - Concurrent")
				AddReportEntry(experiment.Name, experiment)

				mock := newMockTTYSaver(false)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("concurrent-restore", func() {
						done := make(chan bool, 10)
						for i := 0; i < 10; i++ {
							go func() {
								tty.Restore(mock)
								done <- true
							}()
						}
						for i := 0; i < 10; i++ {
							<-done
						}
					})
				}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})
			})
		})
	})

	Describe("SignalHandler Performance", func() {
		Context("handler setup", func() {
			It("should measure SignalHandler setup time", func() {
				experiment := gmeasure.NewExperiment("SignalHandler - Setup")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					mock := newMockTTYSaver(false)
					experiment.MeasureDuration("handler-setup", func() {
						tty.SignalHandler(mock)
					})
				}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})

				// Handler setup should be reasonably fast
				Expect(experiment.Get("handler-setup").Stats().DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 1*time.Millisecond))
			})

			It("should measure SignalHandler with nil", func() {
				experiment := gmeasure.NewExperiment("SignalHandler - Nil")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("handler-nil", func() {
						tty.SignalHandler(nil)
					})
				}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})

				Expect(experiment.Get("handler-nil").Stats().DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 1*time.Millisecond))
			})
		})
	})

	Describe("Mock Operations", func() {
		Context("mock restore performance", func() {
			It("should measure mock Restore method", func() {
				experiment := gmeasure.NewExperiment("Mock - Restore")
				AddReportEntry(experiment.Name, experiment)

				mock := newMockTTYSaver(false)

				experiment.Sample(func(idx int) {
					mock.Reset()
					experiment.MeasureDuration("mock-restore", func() {
						_ = mock.Restore()
					})
				}, gmeasure.SamplingConfig{N: 10000, Duration: 5 * time.Second})

				// Direct mock call should be extremely fast (with race detector it's slower)
				Expect(experiment.Get("mock-restore").Stats().DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 10*time.Microsecond))
			})
		})
	})

	Describe("Stress Performance", func() {
		Context("high volume operations", func() {
			It("should measure rapid Restore calls", func() {
				experiment := gmeasure.NewExperiment("Stress - Rapid Restore")
				AddReportEntry(experiment.Name, experiment)

				mock := newMockTTYSaver(false)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("rapid-restore", func() {
						for i := 0; i < 100; i++ {
							tty.Restore(mock)
						}
					})
				}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})

				// 100 restores should complete quickly
				Expect(experiment.Get("rapid-restore").Stats().DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 1*time.Millisecond))
			})

			It("should measure rapid SignalHandler calls", func() {
				experiment := gmeasure.NewExperiment("Stress - Rapid SignalHandler")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("rapid-handler", func() {
						for i := 0; i < 10; i++ {
							mock := newMockTTYSaver(false)
							tty.SignalHandler(mock)
						}
					})
				}, gmeasure.SamplingConfig{N: 10, Duration: 5 * time.Second})
			})
		})
	})

	Describe("Comparison Benchmarks", func() {
		Context("nil vs valid restore", func() {
			It("should compare nil and valid restore performance", func() {
				experiment := gmeasure.NewExperiment("Comparison - Nil vs Valid")
				AddReportEntry(experiment.Name, experiment)

				mock := newMockTTYSaver(false)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("nil-restore", func() {
						tty.Restore(nil)
					})

					mock.Reset()
					experiment.MeasureDuration("valid-restore", func() {
						tty.Restore(mock)
					})
				}, gmeasure.SamplingConfig{N: 1000, Duration: 5 * time.Second})

				// Both should be fast, nil might be slightly faster
				nilMedian := experiment.Get("nil-restore").Stats().DurationFor(gmeasure.StatMedian)
				validMedian := experiment.Get("valid-restore").Stats().DurationFor(gmeasure.StatMedian)

				Expect(nilMedian).To(BeNumerically("<", 1*time.Microsecond))
				Expect(validMedian).To(BeNumerically("<", 10*time.Microsecond))
			})
		})

		Context("success vs failure restore", func() {
			It("should compare successful and failing restore", func() {
				experiment := gmeasure.NewExperiment("Comparison - Success vs Failure")
				AddReportEntry(experiment.Name, experiment)

				successMock := newMockTTYSaver(false)
				failMock := newMockTTYSaver(true)

				experiment.Sample(func(idx int) {
					successMock.Reset()
					experiment.MeasureDuration("success-restore", func() {
						tty.Restore(successMock)
					})

					failMock.Reset()
					experiment.MeasureDuration("fail-restore", func() {
						tty.Restore(failMock)
					})
				}, gmeasure.SamplingConfig{N: 1000, Duration: 5 * time.Second})

				// Both should be similarly fast
				successMedian := experiment.Get("success-restore").Stats().DurationFor(gmeasure.StatMedian)
				failMedian := experiment.Get("fail-restore").Stats().DurationFor(gmeasure.StatMedian)

				Expect(successMedian).To(BeNumerically("<", 10*time.Microsecond))
				Expect(failMedian).To(BeNumerically("<", 10*time.Microsecond))
			})
		})
	})
})
