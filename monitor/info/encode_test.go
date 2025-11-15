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
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/monitor/info"
)

var _ = Describe("Info Encoding and Marshaling", func() {
	var i info.Info

	BeforeEach(func() {
		var err error
		i, err = info.New("test-service")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("MarshalText", func() {
		Context("with only default name", func() {
			It("should marshal to text successfully", func() {
				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				Expect(text).NotTo(BeEmpty())
				Expect(string(text)).To(Equal("test-service"))
			})
		})

		Context("with registered name function", func() {
			It("should marshal the dynamic name", func() {
				i.RegisterName(func() (string, error) {
					return "dynamic-service", nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(text)).To(Equal("dynamic-service"))
			})
		})

		Context("with registered info function", func() {
			It("should include info in the marshaled text", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"version": "1.0.0",
						"status":  "running",
					}, nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				textStr := string(text)
				Expect(textStr).To(ContainSubstring("test-service"))
				// Info should be included
				Expect(textStr).To(MatchRegexp("version.*1.0.0"))
				Expect(textStr).To(MatchRegexp("status.*running"))
			})

			It("should format info as key-value pairs", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"key1": "value1",
					}, nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				textStr := string(text)
				Expect(textStr).To(ContainSubstring("("))
				Expect(textStr).To(ContainSubstring(")"))
			})
		})

		Context("with special characters", func() {
			It("should handle newlines in info", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"multiline": "line1\nline2",
					}, nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				textStr := string(text)
				// Newlines should be removed or replaced with spaces
				Expect(textStr).NotTo(ContainSubstring("\n"))
			})

			It("should handle carriage returns in info", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"text": "content\r\nmore",
					}, nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				textStr := string(text)
				// Carriage returns should be removed
				Expect(textStr).NotTo(ContainSubstring("\r"))
			})
		})
	})

	Describe("MarshalJSON", func() {
		Context("with only default name", func() {
			It("should marshal to JSON successfully", func() {
				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				Expect(jsonBytes).NotTo(BeEmpty())

				// Parse JSON to verify structure
				var result map[string]interface{}
				err = json.Unmarshal(jsonBytes, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveKey("Name"))
				Expect(result["Name"]).To(Equal("test-service"))
			})

			It("should have Info as empty or null", func() {
				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(jsonBytes, &result)
				Expect(err).NotTo(HaveOccurred())

				// Info should be present but might be null or empty
				if info, ok := result["Info"]; ok && info != nil {
					infoMap := info.(map[string]interface{})
					Expect(infoMap).To(BeEmpty())
				}
			})
		})

		Context("with registered name function", func() {
			It("should marshal the dynamic name to JSON", func() {
				i.RegisterName(func() (string, error) {
					return "json-service", nil
				})

				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(jsonBytes, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result["Name"]).To(Equal("json-service"))
			})
		})

		Context("with registered info function", func() {
			It("should include info in JSON", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"version":     "2.0.0",
						"environment": "production",
					}, nil
				})

				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(jsonBytes, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveKey("Info"))

				infoMap := result["Info"].(map[string]interface{})
				Expect(infoMap).To(HaveKey("version"))
				Expect(infoMap["version"]).To(Equal("2.0.0"))
				Expect(infoMap).To(HaveKey("environment"))
				Expect(infoMap["environment"]).To(Equal("production"))
			})

			It("should handle nested structures in JSON", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"config": map[string]interface{}{
							"port": float64(8080),
							"host": "localhost",
						},
					}, nil
				})

				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(jsonBytes, &result)
				Expect(err).NotTo(HaveOccurred())

				infoMap := result["Info"].(map[string]interface{})
				Expect(infoMap).To(HaveKey("config"))
				config := infoMap["config"].(map[string]interface{})
				Expect(config["port"]).To(Equal(float64(8080)))
				Expect(config["host"]).To(Equal("localhost"))
			})

			It("should handle arrays in JSON", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"tags": []string{"web", "api", "v1"},
					}, nil
				})

				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(jsonBytes, &result)
				Expect(err).NotTo(HaveOccurred())

				infoMap := result["Info"].(map[string]interface{})
				Expect(infoMap).To(HaveKey("tags"))
			})
		})

		Context("JSON validity", func() {
			It("should produce valid JSON", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"key": "value",
					}, nil
				})

				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())

				// Should be parsable JSON
				Expect(json.Valid(jsonBytes)).To(BeTrue())
			})

			It("should handle special JSON characters", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"quote":     `"quoted"`,
						"backslash": `\path\to\file`,
						"unicode":   "unicode: \u2764",
					}, nil
				})

				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				Expect(json.Valid(jsonBytes)).To(BeTrue())

				var result map[string]interface{}
				err = json.Unmarshal(jsonBytes, &result)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Encoding interface compliance", func() {
		It("should implement encoding.TextMarshaler", func() {
			_, err := i.MarshalText()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should implement json.Marshaler", func() {
			_, err := i.MarshalJSON()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should work with json.Marshal", func() {
			i.RegisterName(func() (string, error) {
				return "marshaled-service", nil
			})
			i.RegisterInfo(func() (map[string]interface{}, error) {
				return map[string]interface{}{"test": true}, nil
			})

			// Using standard json.Marshal should work
			jsonBytes, err := json.Marshal(i)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBytes).NotTo(BeEmpty())

			var result map[string]interface{}
			err = json.Unmarshal(jsonBytes, &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["Name"]).To(Equal("marshaled-service"))
		})
	})

	Describe("Edge cases", func() {
		Context("with empty info map", func() {
			It("should handle empty info correctly in text", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{}, nil
				})

				text, err := i.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				// Should just show the name without parentheses or with empty parentheses
				Expect(string(text)).To(Equal("test-service"))
			})

			It("should handle empty info correctly in JSON", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{}, nil
				})

				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(jsonBytes, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result["Name"]).To(Equal("test-service"))
			})
		})

		Context("with nil values in info", func() {
			It("should handle nil values in JSON", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"nil_value": nil,
						"key":       "value",
					}, nil
				})

				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(jsonBytes, &result)
				Expect(err).NotTo(HaveOccurred())

				infoMap := result["Info"].(map[string]interface{})
				Expect(infoMap).To(HaveKey("nil_value"))
				Expect(infoMap["nil_value"]).To(BeNil())
			})
		})

		Context("with large info data", func() {
			It("should handle large info maps", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					largeMap := make(map[string]interface{})
					for n := 0; n < 1000; n++ {
						largeMap[string(rune('a'+n%26))+string(rune('0'+n%10))] = n
					}
					return largeMap, nil
				})

				jsonBytes, err := i.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(jsonBytes)).To(BeNumerically(">", 100))

				var result map[string]interface{}
				err = json.Unmarshal(jsonBytes, &result)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
