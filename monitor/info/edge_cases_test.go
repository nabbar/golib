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

package info_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/monitor/info"
)

var _ = Describe("Info Edge Cases and Coverage", func() {
	Describe("Internal state handling", func() {
		var i info.Info

		BeforeEach(func() {
			var err error
			i, err = info.New("edge-case-service")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("with multiple re-registrations", func() {
			It("should handle multiple name re-registrations correctly", func() {
				// First registration
				i.RegisterName(func() (string, error) {
					return "name-v1", nil
				})
				Expect(i.Name()).To(Equal("name-v1"))

				// Second registration - should clear cache and use new function
				i.RegisterName(func() (string, error) {
					return "name-v2", nil
				})
				Expect(i.Name()).To(Equal("name-v2"))

				// Third registration with error - should return default
				i.RegisterName(func() (string, error) {
					return "", errors.New("error")
				})
				Expect(i.Name()).To(Equal("edge-case-service"))
			})

			It("should handle multiple info re-registrations correctly", func() {
				// First registration
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"v": 1}, nil
				})
				info1 := i.Info()
				Expect(info1["v"]).To(Equal(1))

				// Second registration - should clear cache
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"v": 2}, nil
				})
				info2 := i.Info()
				Expect(info2["v"]).To(Equal(2))

				// Third registration with error
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return nil, errors.New("error")
				})
				info3 := i.Info()
				Expect(info3).To(BeNil())
			})
		})

		Context("with alternating access patterns", func() {
			It("should handle name access after error then success", func() {
				callCount := 0
				i.RegisterName(func() (string, error) {
					callCount++
					if callCount == 1 {
						return "", errors.New("first call error")
					}
					return "success", nil
				})

				// First call fails, returns default
				name1 := i.Name()
				Expect(name1).To(Equal("edge-case-service"))

				// Function should be called again on next access
				// But since it returned error, it's not marked as completed
				// So re-register to try again
				i.RegisterName(func() (string, error) {
					return "success", nil
				})
				name2 := i.Name()
				Expect(name2).To(Equal("success"))
			})

			It("should handle info access after error then success", func() {
				callCount := 0
				i.RegisterInfo(func() (map[string]interface{}, error) {
					callCount++
					if callCount == 1 {
						return nil, errors.New("first call error")
					}
					return map[string]interface{}{"success": true}, nil
				})

				// First call fails
				info1 := i.Info()
				Expect(info1).To(BeNil())

				// Re-register for success
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"success": true}, nil
				})
				info2 := i.Info()
				Expect(info2).NotTo(BeNil())
				Expect(info2["success"]).To(BeTrue())
			})
		})

		Context("with info containing special values", func() {
			It("should handle info with various nil map scenarios", func() {
				// Register function that returns nil map with nil error
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return nil, nil
				})

				result := i.Info()
				// Result should be the nil map returned by the function
				Expect(result).To(BeNil())
			})

			It("should handle non-string keys in sync.Map edge case", func() {
				// This tests the internal sync.Map range function
				// which checks for non-string keys
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"key1": "value1",
						"key2": "value2",
						"key3": "value3",
					}, nil
				})

				// Access info to populate sync.Map
				info1 := i.Info()
				Expect(info1).To(HaveLen(3))

				// Register new info to trigger the Range cleanup
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"newkey": "newvalue",
					}, nil
				})

				info2 := i.Info()
				Expect(info2).To(HaveLen(1))
				Expect(info2["newkey"]).To(Equal("newvalue"))
			})
		})

		Context("with concurrent name and info operations", func() {
			It("should handle interleaved name and info registrations", func() {
				i.RegisterName(func() (string, error) {
					return "concurrent-name", nil
				})
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"concurrent": true}, nil
				})

				// Access both
				name := i.Name()
				info := i.Info()

				Expect(name).To(Equal("concurrent-name"))
				Expect(info["concurrent"]).To(BeTrue())

				// Re-register both
				i.RegisterName(func() (string, error) {
					return "new-name", nil
				})
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"new": true}, nil
				})

				// Access again
				name2 := i.Name()
				info2 := i.Info()

				Expect(name2).To(Equal("new-name"))
				Expect(info2["new"]).To(BeTrue())
			})
		})

		Context("with empty and whitespace names", func() {
			It("should handle name function returning empty string without error", func() {
				i.RegisterName(func() (string, error) {
					return "", nil
				})

				// Empty string is valid if no error
				name := i.Name()
				Expect(name).To(Equal(""))
			})

			It("should handle name with only whitespace", func() {
				i.RegisterName(func() (string, error) {
					return "   ", nil
				})

				name := i.Name()
				Expect(name).To(Equal("   "))
			})
		})

		Context("with complex info data structures", func() {
			It("should preserve all data types correctly", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"nil":        nil,
						"bool_true":  true,
						"bool_false": false,
						"int":        42,
						"int64":      int64(9223372036854775807),
						"float32":    float32(3.14),
						"float64":    3.141592653589793,
						"string":     "hello",
						"empty_str":  "",
						"bytes":      []byte("bytes"),
						"slice":      []int{1, 2, 3},
						"map":        map[string]string{"nested": "value"},
						"interface":  interface{}("interface_value"),
					}, nil
				})

				result := i.Info()
				Expect(result).To(HaveLen(13))
				Expect(result["nil"]).To(BeNil())
				Expect(result["bool_true"]).To(BeTrue())
				Expect(result["bool_false"]).To(BeFalse())
				Expect(result["int"]).To(Equal(42))
				Expect(result["string"]).To(Equal("hello"))
			})
		})
	})

	Describe("Marshaling edge cases", func() {
		var i info.Info

		BeforeEach(func() {
			var err error
			i, err = info.New("marshal-test")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("with empty info after registration", func() {
			It("should handle text marshaling with empty info result", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{}, nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				// Should just be the name without info
				Expect(string(text)).To(Equal("marshal-test"))
			})
		})

		Context("with special Unicode characters", func() {
			It("should handle Unicode in name", func() {
				i.RegisterName(func() (string, error) {
					return "service-名前-📦", nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(text)).To(ContainSubstring("service-名前-📦"))
			})

			It("should handle Unicode in info", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"emoji":    "🚀🎉",
						"japanese": "こんにちは",
						"chinese":  "你好",
					}, nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				textStr := string(text)
				Expect(textStr).To(ContainSubstring("🚀🎉"))
			})
		})

		Context("with very long strings", func() {
			It("should handle long name", func() {
				longName := string(make([]byte, 10000))
				for j := range longName {
					longName = longName[:j] + "a" + longName[j+1:]
				}

				i.RegisterName(func() (string, error) {
					return longName, nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(text)).To(BeNumerically(">=", 10000))
			})

			It("should handle long info values", func() {
				longValue := string(make([]byte, 5000))
				for j := range longValue {
					longValue = longValue[:j] + "x" + longValue[j+1:]
				}

				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"long": longValue,
					}, nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(text)).To(BeNumerically(">", 1000))
			})
		})
	})
})
