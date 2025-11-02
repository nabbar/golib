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
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/monitor/info"
)

var _ = Describe("Info Constructor and Interface", func() {
	Describe("New constructor", func() {
		Context("with valid default name", func() {
			It("should create a new Info instance successfully", func() {
				i, err := info.New("test-service")
				Expect(err).NotTo(HaveOccurred())
				Expect(i).NotTo(BeNil())
			})

			It("should return the default name", func() {
				i, err := info.New("default-service")
				Expect(err).NotTo(HaveOccurred())
				Expect(i.Name()).To(Equal("default-service"))
			})
		})

		Context("with empty default name", func() {
			It("should return an error", func() {
				i, err := info.New("")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("default name cannot be empty"))
				Expect(i).To(BeNil())
			})
		})
	})

	Describe("Name method", func() {
		var i info.Info

		BeforeEach(func() {
			var err error
			i, err = info.New("test-service")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("without registered name function", func() {
			It("should return the default name", func() {
				Expect(i.Name()).To(Equal("test-service"))
			})
		})

		Context("with registered name function", func() {
			It("should return the name from the function", func() {
				i.RegisterName(func() (string, error) {
					return "dynamic-service", nil
				})
				Expect(i.Name()).To(Equal("dynamic-service"))
			})

			It("should return default name if function returns error", func() {
				i.RegisterName(func() (string, error) {
					return "", errors.New("name error")
				})
				Expect(i.Name()).To(Equal("test-service"))
			})

			It("should cache the name after successful call", func() {
				callCount := 0
				i.RegisterName(func() (string, error) {
					callCount++
					return "cached-service", nil
				})

				// First call should invoke the function
				Expect(i.Name()).To(Equal("cached-service"))
				Expect(callCount).To(Equal(1))

				// Second call should use cached value
				Expect(i.Name()).To(Equal("cached-service"))
				Expect(callCount).To(Equal(1))
			})

			It("should allow re-registration and clear cache", func() {
				i.RegisterName(func() (string, error) {
					return "first-service", nil
				})
				Expect(i.Name()).To(Equal("first-service"))

				// Re-register with new function
				i.RegisterName(func() (string, error) {
					return "second-service", nil
				})
				Expect(i.Name()).To(Equal("second-service"))
			})
		})
	})

	Describe("Info method", func() {
		var i info.Info

		BeforeEach(func() {
			var err error
			i, err = info.New("test-service")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("without registered info function", func() {
			It("should return nil", func() {
				result := i.Info()
				Expect(result).To(BeNil())
			})
		})

		Context("with registered info function", func() {
			It("should return the info from the function", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"version": "1.0.0",
						"status":  "running",
					}, nil
				})

				result := i.Info()
				Expect(result).NotTo(BeNil())
				Expect(result).To(HaveKey("version"))
				Expect(result["version"]).To(Equal("1.0.0"))
				Expect(result).To(HaveKey("status"))
				Expect(result["status"]).To(Equal("running"))
			})

			It("should return nil if function returns error", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return nil, errors.New("info error")
				})
				result := i.Info()
				Expect(result).To(BeNil())
			})

			It("should cache the info after successful call", func() {
				callCount := 0
				i.RegisterInfo(func() (map[string]interface{}, error) {
					callCount++
					return map[string]interface{}{
						"cached": true,
					}, nil
				})

				// First call should invoke the function
				result1 := i.Info()
				Expect(result1).NotTo(BeNil())
				Expect(callCount).To(Equal(1))

				// Second call should use cached value
				result2 := i.Info()
				Expect(result2).NotTo(BeNil())
				Expect(callCount).To(Equal(1))
			})

			It("should allow re-registration and clear cache", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"version": "1.0.0"}, nil
				})
				result1 := i.Info()
				Expect(result1["version"]).To(Equal("1.0.0"))

				// Re-register with new function
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"version": "2.0.0"}, nil
				})
				result2 := i.Info()
				Expect(result2["version"]).To(Equal("2.0.0"))
			})

			It("should handle empty map from function", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{}, nil
				})
				result := i.Info()
				Expect(result).NotTo(BeNil())
				Expect(result).To(BeEmpty())
			})

			It("should handle various data types in info", func() {
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"string": "value",
						"int":    42,
						"float":  3.14,
						"bool":   true,
						"array":  []string{"a", "b", "c"},
						"nested": map[string]interface{}{"key": "value"},
					}, nil
				})

				result := i.Info()
				Expect(result).NotTo(BeNil())
				Expect(result["string"]).To(Equal("value"))
				Expect(result["int"]).To(Equal(42))
				Expect(result["float"]).To(Equal(3.14))
				Expect(result["bool"]).To(BeTrue())
				Expect(result["array"]).To(Equal([]string{"a", "b", "c"}))
				Expect(result["nested"]).To(Equal(map[string]interface{}{"key": "value"}))
			})
		})
	})

	Describe("Interface compliance", func() {
		It("should implement montps.Info interface", func() {
			i, err := info.New("test-service")
			Expect(err).NotTo(HaveOccurred())

			// Verify it has Name() method
			Expect(i.Name()).To(Equal("test-service"))

			// Verify it has Info() method
			Expect(i.Info()).To(BeNil())
		})
	})
})

var _ = Describe("Info Concurrent Access", func() {
	var i info.Info

	BeforeEach(func() {
		var err error
		i, err = info.New("concurrent-test")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should handle concurrent RegisterName calls safely", func() {
		var wg sync.WaitGroup
		iterations := 100

		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				i.RegisterName(func() (string, error) {
					return "concurrent-service", nil
				})
			}(n)
		}

		wg.Wait()
		// Should not panic and should return a valid name
		Expect(i.Name()).NotTo(BeEmpty())
	})

	It("should handle concurrent RegisterInfo calls safely", func() {
		var wg sync.WaitGroup
		iterations := 100

		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"index": index}, nil
				})
			}(n)
		}

		wg.Wait()
		// Should not panic
		result := i.Info()
		Expect(result).NotTo(BeNil())
	})

	It("should handle concurrent Name calls safely", func() {
		i.RegisterName(func() (string, error) {
			return "concurrent-service", nil
		})

		var wg sync.WaitGroup
		iterations := 100
		results := make([]string, iterations)

		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				results[index] = i.Name()
			}(n)
		}

		wg.Wait()

		// All results should be consistent
		for _, name := range results {
			Expect(name).NotTo(BeEmpty())
		}
	})

	It("should handle concurrent Info calls safely", func() {
		i.RegisterInfo(func() (map[string]interface{}, error) {
			return map[string]interface{}{"concurrent": true}, nil
		})

		var wg sync.WaitGroup
		iterations := 100

		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result := i.Info()
				Expect(result).NotTo(BeNil())
			}()
		}

		wg.Wait()
	})

	It("should handle mixed concurrent operations safely", func() {
		var wg sync.WaitGroup
		iterations := 50

		// Concurrent RegisterName
		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				i.RegisterName(func() (string, error) {
					return "mixed-service", nil
				})
			}()
		}

		// Concurrent RegisterInfo
		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"mixed": true}, nil
				})
			}()
		}

		// Concurrent Name calls
		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = i.Name()
			}()
		}

		// Concurrent Info calls
		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = i.Info()
			}()
		}

		wg.Wait()
		// Should complete without panic
		Expect(i.Name()).NotTo(BeEmpty())
	})
})
