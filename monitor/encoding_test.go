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

package monitor_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync/atomic"
	"time"

	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor Encoding", func() {
	var (
		ctx context.Context
		cnl context.CancelFunc
		nfo montps.Info
		mon montps.Monitor
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, 5*time.Second)
		nfo = newInfo(func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"version": "1.0.0",
				"env":     "test",
			}, nil
		})
		mon = newMonitor(x, nfo)
		Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if mon != nil && mon.IsRunning() {
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		}
		if cnl != nil {
			cnl()
		}
	})

	Describe("MarshalText", func() {
		It("should marshal monitor to text format", func() {
			mon.InfoUpd(newInfoWithName("test-encoding", func() (map[string]interface{}, error) {
				return map[string]interface{}{
					"msg": "encoding-test",
				}, nil
			}))

			text, err := mon.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(text).ToNot(BeEmpty())

			textStr := string(text)
			Expect(textStr).To(ContainSubstring("encoding-test"))
		})

		It("should include status in text output", func() {
			text, err := mon.MarshalText()
			Expect(err).ToNot(HaveOccurred())

			textStr := string(text)
			// Should contain a status (KO, Warn, or OK)
			Expect(textStr).To(SatisfyAny(
				ContainSubstring("KO"),
				ContainSubstring("Warn"),
				ContainSubstring("OK"),
			))
		})

		It("should include timing information", func() {
			// Run a health check first
			called := new(atomic.Bool)
			mon.SetHealthCheck(func(ctx context.Context) error {
				called.Store(true)
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
			Expect(mon.IsRunning()).To(BeTrue())
			time.Sleep(100 * time.Millisecond)
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
			Expect(called.Load()).To(BeTrue())

			text, err := mon.MarshalText()
			Expect(err).ToNot(HaveOccurred())

			textStr := string(text)
			// Should contain duration separators
			Expect(textStr).To(ContainSubstring("/"))
		})

		It("should include info metadata in text output", func() {
			// Run a health check first
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})
			i := newInfoWithName("encoding-test", nil)
			mon.InfoUpd(i)
			Expect(mon.SetConfig(ctx, newConfig(i))).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())

			text, err := mon.MarshalText()
			Expect(err).ToNot(HaveOccurred())

			textStr := string(text)
			// Info should be included in parentheses
			Expect(textStr).To(ContainSubstring("encoding-test"))
		})

		It("should include error message when health check fails", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return ErrorMockTest
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(150 * time.Millisecond)
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())

			text, err := mon.MarshalText()
			Expect(err).ToNot(HaveOccurred())

			textStr := string(text)
			Expect(textStr).To(ContainSubstring("mock test error"))
		})
	})

	Describe("MarshalJSON", func() {
		It("should marshal monitor to JSON format", func() {
			jsonData, err := mon.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(jsonData).ToNot(BeEmpty())

			var result map[string]interface{}
			Expect(json.Unmarshal(jsonData, &result)).ToNot(HaveOccurred())
		})

		It("should include status in JSON output", func() {
			jsonData, err := mon.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]interface{}
			Expect(json.Unmarshal(jsonData, &result)).ToNot(HaveOccurred())

			Expect(result).To(HaveKey("Status"))
		})

		It("should include name in JSON output", func() {
			i := newInfoWithName("encoding-test", nil)
			mon.InfoUpd(i)

			Expect(mon.SetConfig(ctx, newConfig(i))).ToNot(HaveOccurred())

			jsonData, err := mon.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]interface{}
			Expect(json.Unmarshal(jsonData, &result)).ToNot(HaveOccurred())

			Expect(result).To(HaveKey("Name"))
			Expect(result["Name"]).To(Equal("encoding-test"))
		})

		It("should include timing metrics in JSON output", func() {
			// Run a health check first
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())

			jsonData, err := mon.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]interface{}
			Expect(json.Unmarshal(jsonData, &result)).ToNot(HaveOccurred())

			Expect(result).To(HaveKey("Latency"))
			Expect(result).To(HaveKey("Uptime"))
			Expect(result).To(HaveKey("Downtime"))
		})

		It("should include message when present", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return ErrorMockTest
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(150 * time.Millisecond)
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())

			jsonData, err := mon.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]interface{}
			Expect(json.Unmarshal(jsonData, &result)).ToNot(HaveOccurred())

			Expect(result).To(HaveKey("Message"))
			message, ok := result["Message"].(string)
			Expect(ok).To(BeTrue())
			Expect(message).To(ContainSubstring("mock test error"))
		})

		It("should have valid JSON structure", func() {
			jsonData, err := mon.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			// Should be valid JSON
			Expect(json.Valid(jsonData)).To(BeTrue())

			// Should be a JSON object
			Expect(string(jsonData)).To(HavePrefix("{"))
			Expect(string(jsonData)).To(HaveSuffix("}"))
		})
	})

	Describe("Text Format Structure", func() {
		It("should follow expected text format pattern", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())

			text, err := mon.MarshalText()
			Expect(err).ToNot(HaveOccurred())

			textStr := string(text)

			// Should contain status separator ": "
			Expect(textStr).To(ContainSubstring(": "))

			// Should contain part separator " | "
			Expect(textStr).To(ContainSubstring(" | "))

			// Should contain time separator " / "
			Expect(textStr).To(ContainSubstring(" / "))
		})

		It("should start with status", func() {
			text, err := mon.MarshalText()
			Expect(err).ToNot(HaveOccurred())

			textStr := string(text)
			parts := strings.Split(textStr, ": ") // KO: <monitor name> (<info name> (<data name>: <data value>, <data name>: <data value>))
			Expect(parts).To(HaveLen(4))

			// First part should be a valid status
			status := parts[0]
			Expect(status).To(SatisfyAny(
				Equal("KO"),
				Equal("Warn"),
				Equal("OK"),
			))
		})
	})
})
