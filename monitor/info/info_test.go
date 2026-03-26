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

// This suite tests the core functionality of the Info component,
// including its constructor, the Name() and Info() methods, and basic
// registration behavior.
var _ = Describe("Info Constructor and Interface", func() {
	// Tests for the New() constructor function.
	Describe("New constructor", func() {
		Context("with valid default name", func() {
			// TC-CORE-001
			It("should create a new Info instance successfully", func() {
				i, err := info.New("test-service")
				Expect(err).NotTo(HaveOccurred())
				Expect(i).NotTo(BeNil())
			})

			// TC-CORE-002
			It("should return the default name", func() {
				i, err := info.New("default-service")
				Expect(err).NotTo(HaveOccurred())
				Expect(i.Name()).To(Equal("default-service"))
			})
		})

		Context("with empty default name", func() {
			// TC-CORE-003
			It("should return an error", func() {
				i, err := info.New("")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("default name cannot be empty"))
				Expect(i).To(BeNil())
			})
		})
	})

	// Tests for the Name() method behavior.
	Describe("Name method", func() {
		var i info.Info

		BeforeEach(func() {
			var err error
			i, err = info.New("test-service")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("without registered name function", func() {
			// TC-CORE-004
			It("should return the default name", func() {
				Expect(i.Name()).To(Equal("test-service"))
			})
		})

		Context("with registered name function", func() {
			// TC-CORE-005
			It("should return the name from the function", func() {
				i.RegisterName(func() (string, error) {
					return "dynamic-service", nil
				})
				Expect(i.Name()).To(Equal("dynamic-service"))
			})

			// TC-CORE-006
			It("should return default name if function returns error", func() {
				i.RegisterName(func() (string, error) {
					return "", errors.New("name error")
				})
				Expect(i.Name()).To(Equal("test-service"))
			})

			// TC-CORE-007
			// This test verifies that the implementation does NOT cache the name,
			// calling the registered function on each invocation.
			It("should NOT cache the name after successful call", func() {
				callCount := 0
				i.RegisterName(func() (string, error) {
					callCount++
					return "cached-service", nil
				})

				// First call should invoke the function
				Expect(i.Name()).To(Equal("cached-service"))
				Expect(callCount).To(Equal(1))

				// Second call should also invoke the function (no caching)
				Expect(i.Name()).To(Equal("cached-service"))
				Expect(callCount).To(Equal(2))
			})

			// TC-CORE-008
			It("should allow re-registration", func() {
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

	// Tests for the Data() method behavior.
	Describe("Data method", func() {
		var i info.Info

		BeforeEach(func() {
			var err error
			i, err = info.New("test-service")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("without registered info function", func() {
			// TC-CORE-009
			It("should return empty map", func() {
				result := i.Data()
				Expect(result).To(BeEmpty())
			})
		})

		Context("with registered info function", func() {
			// TC-CORE-010
			It("should return the info from the function", func() {
				i.RegisterData(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"version": "1.0.0",
						"status":  "running",
					}, nil
				})

				result := i.Data()
				Expect(result).NotTo(BeNil())
				Expect(result).To(HaveKey("version"))
				Expect(result["version"]).To(Equal("1.0.0"))
				Expect(result).To(HaveKey("status"))
				Expect(result["status"]).To(Equal("running"))
			})

			// TC-CORE-011
			It("should return empty map if function returns error", func() {
				i.RegisterData(func() (map[string]interface{}, error) {
					return nil, errors.New("info error")
				})
				result := i.Data()
				Expect(result).To(BeEmpty())
			})

			// TC-CORE-012
			// This test verifies that the implementation does NOT cache the info data,
			// calling the registered function on each invocation.
			It("should NOT cache the info after successful call", func() {
				callCount := 0
				i.RegisterData(func() (map[string]interface{}, error) {
					callCount++
					return map[string]interface{}{
						"cached": true,
					}, nil
				})

				// First call should invoke the function
				result1 := i.Data()
				Expect(result1).NotTo(BeNil())
				Expect(callCount).To(Equal(1))

				// Second call should also invoke the function (no caching)
				result2 := i.Data()
				Expect(result2).NotTo(BeNil())
				Expect(callCount).To(Equal(2))
			})

			// TC-CORE-013
			It("should allow re-registration", func() {
				i.RegisterData(func() (map[string]interface{}, error) {
					return map[string]interface{}{"version": "1.0.0"}, nil
				})
				result1 := i.Data()
				Expect(result1["version"]).To(Equal("1.0.0"))

				// Re-register with new function
				i.RegisterData(func() (map[string]interface{}, error) {
					return map[string]interface{}{"version": "2.0.0"}, nil
				})
				result2 := i.Data()
				Expect(result2["version"]).To(Equal("2.0.0"))
			})

			// TC-CORE-014
			It("should handle empty map from function", func() {
				i.RegisterData(func() (map[string]interface{}, error) {
					return map[string]interface{}{}, nil
				})
				result := i.Data()
				Expect(result).NotTo(BeNil())
				Expect(result).To(BeEmpty())
			})

			// TC-CORE-015
			It("should handle various data types in info", func() {
				i.RegisterData(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"string": "value",
						"int":    42,
						"float":  3.14,
						"bool":   true,
						"array":  []string{"a", "b", "c"},
						"nested": map[string]interface{}{"key": "value"},
					}, nil
				})

				result := i.Data()
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

	// This test ensures the Info object adheres to the expected interface contract.
	Describe("Interface compliance", func() {
		// TC-CORE-016
		It("should implement montps.Info interface", func() {
			i, err := info.New("test-service")
			Expect(err).NotTo(HaveOccurred())

			// Verify it has Name() method
			Expect(i.Name()).To(Equal("test-service"))

			// Verify it has Data() method
			Expect(i.Data()).To(BeEmpty())
		})
	})
})

// This suite tests the thread-safety of the Info component when its methods
// are called concurrently from multiple goroutines.
var _ = Describe("Info Concurrent Access", func() {
	var i info.Info

	BeforeEach(func() {
		var err error
		i, err = info.New("concurrent-test")
		Expect(err).NotTo(HaveOccurred())
	})

	// TC-CONC-001
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

	// TC-CONC-002
	It("should handle concurrent RegisterData calls safely", func() {
		var wg sync.WaitGroup
		iterations := 100

		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				i.RegisterData(func() (map[string]interface{}, error) {
					return map[string]interface{}{"index": index}, nil
				})
			}(n)
		}

		wg.Wait()
		// Should not panic
		result := i.Data()
		Expect(result).NotTo(BeNil())
	})

	// TC-CONC-003
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

	// TC-CONC-004
	It("should handle concurrent Data calls safely", func() {
		i.RegisterData(func() (map[string]interface{}, error) {
			return map[string]interface{}{"concurrent": true}, nil
		})

		var wg sync.WaitGroup
		iterations := 100

		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result := i.Data()
				Expect(result).NotTo(BeNil())
			}()
		}

		wg.Wait()
	})

	// TC-CONC-005
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

		// Concurrent RegisterData
		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				i.RegisterData(func() (map[string]interface{}, error) {
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

		// Concurrent Data calls
		for n := 0; n < iterations; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = i.Data()
			}()
		}

		wg.Wait()
		// Should complete without panic
		Expect(i.Name()).NotTo(BeEmpty())
	})
})
