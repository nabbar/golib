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

package socket_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gmeasure"

	libsck "github.com/nabbar/golib/socket"
)

var _ = Describe("[TC-BM] Socket Performance Benchmarks", func() {
	Describe("ErrorFilter performance", func() {
		It("[TC-BM-001] should benchmark ErrorFilter with various error types", func() {
			experiment := gmeasure.NewExperiment("ErrorFilter operations")
			AddReportEntry(experiment.Name, experiment)

			normalErr := fmt.Errorf("connection timeout")
			closedErr := fmt.Errorf("use of closed network connection")

			experiment.SampleDuration("Normal error", func(idx int) {
				_ = libsck.ErrorFilter(normalErr)
			}, gmeasure.SamplingConfig{N: 10000, Duration: 0})

			experiment.SampleDuration("Nil error", func(idx int) {
				_ = libsck.ErrorFilter(nil)
			}, gmeasure.SamplingConfig{N: 10000, Duration: 0})

			experiment.SampleDuration("Closed connection error", func(idx int) {
				_ = libsck.ErrorFilter(closedErr)
			}, gmeasure.SamplingConfig{N: 10000, Duration: 0})
		})
	})

	Describe("ConnState String performance", func() {
		It("[TC-BM-002] should benchmark ConnState.String method", func() {
			experiment := gmeasure.NewExperiment("ConnState String conversion")
			AddReportEntry(experiment.Name, experiment)

			states := []libsck.ConnState{
				libsck.ConnectionDial,
				libsck.ConnectionNew,
				libsck.ConnectionRead,
				libsck.ConnectionCloseRead,
				libsck.ConnectionHandler,
				libsck.ConnectionWrite,
				libsck.ConnectionCloseWrite,
				libsck.ConnectionClose,
			}

			for _, state := range states {
				stateName := state.String()
				experiment.SampleDuration(stateName, func(idx int) {
					_ = state.String()
				}, gmeasure.SamplingConfig{N: 10000, Duration: 0})
			}

			experiment.SampleDuration("Unknown state", func(idx int) {
				state := libsck.ConnState(255)
				_ = state.String()
			}, gmeasure.SamplingConfig{N: 10000, Duration: 0})
		})
	})

	Describe("Real-world scenarios", func() {
		It("[TC-BM-003] should benchmark error handling in connection lifecycle", func() {
			experiment := gmeasure.NewExperiment("Connection lifecycle error handling")

			experiment.Sample(func(idx int) {
				errors := []error{
					nil,
					fmt.Errorf("connection timeout"),
					fmt.Errorf("use of closed network connection"),
					fmt.Errorf("connection refused"),
					fmt.Errorf("broken pipe"),
					nil,
				}

				experiment.MeasureDuration("error-lifecycle", func() {
					for _, err := range errors {
						_ = libsck.ErrorFilter(err)
					}
				})
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			AddReportEntry(experiment.Name, experiment)
		})

		It("[TC-BM-004] should benchmark state tracking overhead", func() {
			experiment := gmeasure.NewExperiment("State tracking overhead")

			experiment.Sample(func(idx int) {
				states := []libsck.ConnState{
					libsck.ConnectionDial,
					libsck.ConnectionNew,
					libsck.ConnectionRead,
					libsck.ConnectionHandler,
					libsck.ConnectionWrite,
					libsck.ConnectionClose,
				}

				experiment.MeasureDuration("state-tracking", func() {
					for _, state := range states {
						_ = state.String()
					}
				})
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			AddReportEntry(experiment.Name, experiment)
		})
	})
})
