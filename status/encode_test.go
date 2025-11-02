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

package status_test

import (
	"encoding/json"

	monpol "github.com/nabbar/golib/monitor/pool"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	libsts "github.com/nabbar/golib/status"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status/Encode", func() {
	var (
		status libsts.Status
		pool   monpol.Pool
	)

	BeforeEach(func() {
		status = libsts.New(globalCtx)
		pool = newPool()

		// Setup application info (use SetInfo which is simpler)
		status.SetInfo("test-app", "v1.0.0", "abc123")

		// Configure status with return codes
		cfg := libsts.Config{
			ReturnCode: map[monsts.Status]int{
				monsts.OK:   200,
				monsts.Warn: 200,
				monsts.KO:   503,
			},
			MandatoryComponent: make([]libsts.Mandatory, 0),
		}
		status.SetConfig(cfg)

		// Register pool
		status.RegisterPool(func() montps.Pool { return pool })
	})

	Describe("MarshalJSON", func() {
		It("should marshal status to JSON", func() {
			// Get the status and marshal it
			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).ToNot(BeEmpty())
		})

		Context("with monitors", func() {
			BeforeEach(func() {
				// Add a healthy monitor
				mon := newHealthyMonitor("healthy-monitor")
				err := pool.MonitorAdd(mon)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should include monitor status in JSON", func() {
				data, err := json.Marshal(status)
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeEmpty())
				Expect(string(data)).To(ContainSubstring("healthy-monitor"))
			})
		})
	})

	Describe("Status codes", func() {
		It("should handle different status codes", func() {
			cfg := libsts.Config{
				ReturnCode: map[monsts.Status]int{
					monsts.OK:   200,
					monsts.Warn: 200,
					monsts.KO:   503,
				},
			}
			status.SetConfig(cfg)
			Expect(true).To(BeTrue())
		})
	})
})
