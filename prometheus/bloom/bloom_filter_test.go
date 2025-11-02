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

package bloom_test

import (
	"github.com/nabbar/golib/prometheus/bloom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BloomFilter", func() {
	var bf bloom.BloomFilter

	BeforeEach(func() {
		bf = bloom.NewBloomFilter()
	})

	Describe("NewBloomFilter", func() {
		It("should create a new Bloom filter", func() {
			Expect(bf).ToNot(BeNil())
		})
	})

	Describe("Add and Contains", func() {
		Context("with single value", func() {
			It("should add and retrieve a value", func() {
				bf.Add("test-value")
				Expect(bf.Contains("test-value")).To(BeTrue())
			})

			It("should return false for non-existent value", func() {
				bf.Add("test-value")
				Expect(bf.Contains("different-value")).To(BeFalse())
			})

			It("should handle empty string", func() {
				Expect(bf.Contains("")).To(BeFalse())
			})
		})

		Context("with multiple values", func() {
			It("should handle multiple distinct values", func() {
				values := []string{"value1", "value2", "value3", "value4", "value5"}

				for _, v := range values {
					bf.Add(v)
				}

				for _, v := range values {
					Expect(bf.Contains(v)).To(BeTrue())
				}
			})

			It("should not find values that weren't added", func() {
				bf.Add("exists1")
				bf.Add("exists2")

				Expect(bf.Contains("not-added")).To(BeFalse())
				Expect(bf.Contains("also-not-added")).To(BeFalse())
			})
		})

		Context("with duplicate values", func() {
			It("should handle adding same value multiple times", func() {
				bf.Add("duplicate")
				bf.Add("duplicate")
				bf.Add("duplicate")

				Expect(bf.Contains("duplicate")).To(BeTrue())
			})
		})

		Context("with various string types", func() {
			It("should handle numeric strings", func() {
				bf.Add("12345")
				bf.Add("67890")

				Expect(bf.Contains("12345")).To(BeTrue())
				Expect(bf.Contains("67890")).To(BeTrue())
				Expect(bf.Contains("11111")).To(BeFalse())
			})

			It("should handle special characters", func() {
				bf.Add("test@example.com")
				bf.Add("path/to/resource")
				bf.Add("key=value&another=value")

				Expect(bf.Contains("test@example.com")).To(BeTrue())
				Expect(bf.Contains("path/to/resource")).To(BeTrue())
				Expect(bf.Contains("key=value&another=value")).To(BeTrue())
			})

			It("should handle Unicode characters", func() {
				bf.Add("こんにちは")
				bf.Add("مرحبا")
				bf.Add("🚀🎉")

				Expect(bf.Contains("こんにちは")).To(BeTrue())
				Expect(bf.Contains("مرحبا")).To(BeTrue())
				Expect(bf.Contains("🚀🎉")).To(BeTrue())
			})

			It("should distinguish between similar strings", func() {
				bf.Add("test")
				bf.Add("Test")
				bf.Add("TEST")

				Expect(bf.Contains("test")).To(BeTrue())
				Expect(bf.Contains("Test")).To(BeTrue())
				Expect(bf.Contains("TEST")).To(BeTrue())
				Expect(bf.Contains("tEsT")).To(BeFalse())
			})
		})

		Context("with long strings", func() {
			It("should handle very long strings", func() {
				longString := string(make([]byte, 10000))
				for i := range longString {
					longString = string(append([]byte(longString[:i]), byte('a'+i%26)))
				}

				bf.Add(longString)
				Expect(bf.Contains(longString)).To(BeTrue())
			})
		})
	})

	Describe("False Positive Rate", func() {
		It("should have low false positive rate with reasonable data set", func() {
			// Add 1000 values
			for i := 0; i < 1000; i++ {
				bf.Add(GinkgoT().Name() + "-" + string(rune('0'+i%10)) + string(rune('a'+i%26)))
			}

			// Test 1000 non-existent values
			falsePositives := 0
			for i := 1000; i < 2000; i++ {
				if bf.Contains("not-added-" + string(rune('0'+i%10)) + string(rune('a'+i%26))) {
					falsePositives++
				}
			}

			// False positive rate should be very low (< 5%)
			rate := float64(falsePositives) / 1000.0
			Expect(rate).To(BeNumerically("<", 0.05))
		})
	})

	Describe("Bloom Filter Properties", func() {
		It("should never produce false negatives", func() {
			// Add values and verify they're all found
			testValues := []string{
				"property-test-1",
				"property-test-2",
				"property-test-3",
				"property-test-4",
				"property-test-5",
			}

			for _, v := range testValues {
				bf.Add(v)
			}

			// All added values MUST be found (no false negatives)
			for _, v := range testValues {
				Expect(bf.Contains(v)).To(BeTrue(), "Bloom filter should never produce false negatives for: "+v)
			}
		})
	})

	Describe("Edge Cases", func() {
		It("should handle empty string lookup", func() {
			bf.Add("something")
			Expect(bf.Contains("")).To(BeFalse())
		})

		It("should handle whitespace-only strings", func() {
			bf.Add("   ")
			bf.Add("\t")
			bf.Add("\n")

			Expect(bf.Contains("   ")).To(BeTrue())
			Expect(bf.Contains("\t")).To(BeTrue())
			Expect(bf.Contains("\n")).To(BeTrue())
			Expect(bf.Contains(" ")).To(BeFalse())
		})

		It("should work with single character strings", func() {
			bf.Add("a")
			bf.Add("b")
			bf.Add("1")

			Expect(bf.Contains("a")).To(BeTrue())
			Expect(bf.Contains("b")).To(BeTrue())
			Expect(bf.Contains("1")).To(BeTrue())
			Expect(bf.Contains("c")).To(BeFalse())
		})
	})
})
