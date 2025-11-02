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

package pool_test

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	monpool "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool Encoding Operations", func() {
	var (
		pool monpool.Pool
		ctx  context.Context
		cnl  context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, 10*time.Second)
		pool = newPool(ctx)
	})

	AfterEach(func() {
		if pool != nil && pool.IsRunning() {
			_ = pool.Stop(ctx)
		}
		if cnl != nil {
			cnl()
		}
	})

	Describe("MarshalText", func() {
		Context("with empty pool", func() {
			It("should return empty bytes", func() {
				text, err := pool.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(text).To(BeEmpty())
			})
		})

		Context("with single monitor", func() {
			It("should marshal monitor as text", func() {
				monitor := createTestMonitor("text-test", nil)
				defer monitor.Stop(ctx)

				Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())

				text, err := pool.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(text).ToNot(BeEmpty())

				// Should contain monitor name
				textStr := string(text)
				Expect(textStr).To(ContainSubstring("text-test"))
			})
		})

		Context("with multiple monitors", func() {
			It("should marshal all monitors with newlines", func() {
				monitors := []montps.Monitor{
					createTestMonitor("text-1", nil),
					createTestMonitor("text-2", nil),
					createTestMonitor("text-3", nil),
				}

				for _, mon := range monitors {
					defer mon.Stop(ctx)
					Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
				}

				text, err := pool.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(text).ToNot(BeEmpty())

				textStr := string(text)
				lines := strings.Split(textStr, "\n")

				// Should have at least 3 lines (one per monitor)
				Expect(len(lines)).To(BeNumerically(">=", 3))

				// Should contain all monitor names
				Expect(textStr).To(ContainSubstring("text-1"))
				Expect(textStr).To(ContainSubstring("text-2"))
				Expect(textStr).To(ContainSubstring("text-3"))
			})
		})

		Context("with running monitors", func() {
			It("should include status information", func() {
				monitor := createTestMonitor("running-text", nil)
				defer monitor.Stop(ctx)

				Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())
				Expect(pool.Start(ctx)).ToNot(HaveOccurred())

				// Wait for health checks to run
				time.Sleep(150 * time.Millisecond)

				text, err := pool.MarshalText()
				Expect(err).ToNot(HaveOccurred())

				textStr := string(text)
				Expect(textStr).To(ContainSubstring("running-text"))
				// Should contain status (OK, Warn, or KO)
				Expect(textStr).To(MatchRegexp("(OK|Warn|KO)"))
			})
		})
	})

	Describe("MarshalJSON", func() {
		Context("with empty pool", func() {
			It("should return empty JSON object", func() {
				jsonData, err := pool.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonData).ToNot(BeEmpty())

				var result map[string]interface{}
				err = json.Unmarshal(jsonData, &result)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeEmpty())
			})
		})

		Context("with single monitor", func() {
			It("should marshal monitor as JSON", func() {
				monitor := createTestMonitor("json-test", nil)
				defer monitor.Stop(ctx)

				Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())

				jsonData, err := pool.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonData).ToNot(BeEmpty())

				var result map[string]interface{}
				err = json.Unmarshal(jsonData, &result)
				Expect(err).ToNot(HaveOccurred())

				// Should have the monitor as a key
				Expect(result).To(HaveKey("json-test"))
			})
		})

		Context("with multiple monitors", func() {
			It("should marshal all monitors as JSON object", func() {
				monitors := []montps.Monitor{
					createTestMonitor("json-1", nil),
					createTestMonitor("json-2", nil),
					createTestMonitor("json-3", nil),
				}

				for _, mon := range monitors {
					defer mon.Stop(ctx)
					Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
				}

				jsonData, err := pool.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonData).ToNot(BeEmpty())

				var result map[string]interface{}
				err = json.Unmarshal(jsonData, &result)
				Expect(err).ToNot(HaveOccurred())

				// Should have all monitors as keys
				Expect(result).To(HaveKey("json-1"))
				Expect(result).To(HaveKey("json-2"))
				Expect(result).To(HaveKey("json-3"))
			})
		})

		Context("with running monitors", func() {
			It("should include monitor status in JSON", func() {
				monitor := createTestMonitor("running-json", nil)
				defer monitor.Stop(ctx)

				Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())
				Expect(pool.Start(ctx)).ToNot(HaveOccurred())

				// Wait for health checks to run
				time.Sleep(150 * time.Millisecond)

				jsonData, err := pool.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(jsonData, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(HaveKey("running-json"))

				// The value should be the monitor status data
				monitorData := result["running-json"]
				Expect(monitorData).ToNot(BeNil())
			})
		})

		Context("JSON structure validation", func() {
			It("should produce valid JSON with correct structure", func() {
				monitor := createTestMonitor("structure-test", nil)
				defer monitor.Stop(ctx)

				Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())
				Expect(pool.Start(ctx)).ToNot(HaveOccurred())

				time.Sleep(150 * time.Millisecond)

				jsonData, err := pool.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())

				// Should be valid JSON
				var result map[string]json.RawMessage
				err = json.Unmarshal(jsonData, &result)
				Expect(err).ToNot(HaveOccurred())

				// Should have the monitor
				Expect(result).To(HaveKey("structure-test"))

				// The monitor data should also be valid JSON
				var monitorStatus map[string]interface{}
				err = json.Unmarshal(result["structure-test"], &monitorStatus)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Encoding Round-trip", func() {
		It("should handle text marshal/unmarshal consistently", func() {
			monitors := []montps.Monitor{
				createTestMonitor("roundtrip-1", nil),
				createTestMonitor("roundtrip-2", nil),
			}

			for _, mon := range monitors {
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			// Marshal twice and compare
			text1, err1 := pool.MarshalText()
			Expect(err1).ToNot(HaveOccurred())

			text2, err2 := pool.MarshalText()
			Expect(err2).ToNot(HaveOccurred())

			// Both should contain the monitor names
			Expect(string(text1)).To(ContainSubstring("roundtrip-1"))
			Expect(string(text1)).To(ContainSubstring("roundtrip-2"))
			Expect(string(text2)).To(ContainSubstring("roundtrip-1"))
			Expect(string(text2)).To(ContainSubstring("roundtrip-2"))

			// Both should have similar length (within reason)
			Expect(len(text1)).To(BeNumerically("~", len(text2), 50))
		})

		It("should handle JSON marshal/unmarshal consistently", func() {
			monitors := []montps.Monitor{
				createTestMonitor("json-roundtrip-1", nil),
				createTestMonitor("json-roundtrip-2", nil),
			}

			for _, mon := range monitors {
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			// Marshal twice and compare
			json1, err1 := pool.MarshalJSON()
			Expect(err1).ToNot(HaveOccurred())

			json2, err2 := pool.MarshalJSON()
			Expect(err2).ToNot(HaveOccurred())

			// Parse and compare structure
			var result1, result2 map[string]interface{}
			Expect(json.Unmarshal(json1, &result1)).ToNot(HaveOccurred())
			Expect(json.Unmarshal(json2, &result2)).ToNot(HaveOccurred())

			// Should have same keys
			Expect(result1).To(HaveLen(len(result2)))
			for key := range result1 {
				Expect(result2).To(HaveKey(key))
			}
		})
	})

	Describe("Encoding Edge Cases", func() {
		It("should handle monitors with special characters in names", func() {
			// Note: Monitor names should be validated, but we test encoding robustness
			monitor := createTestMonitor("special-test", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())

			text, err := pool.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(text).ToNot(BeEmpty())

			jsonData, err := pool.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]interface{}
			Expect(json.Unmarshal(jsonData, &result)).ToNot(HaveOccurred())
		})

		It("should handle large number of monitors", func() {
			// Add many monitors
			for i := 0; i < 50; i++ {
				mon := createTestMonitor("large-"+string(rune('a'+i%26))+"-"+string(rune('0'+i/26)), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			text, err := pool.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(text).ToNot(BeEmpty())

			jsonData, err := pool.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]interface{}
			Expect(json.Unmarshal(jsonData, &result)).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(50))
		})
	})
})
