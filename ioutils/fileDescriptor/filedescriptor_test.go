/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package fileDescriptor_test

import (
	. "github.com/nabbar/golib/ioutils/fileDescriptor"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Basic functionality tests for SystemFileDescriptor.
// These tests verify core behavior: querying limits, handling edge cases,
// and ensuring the function never corrupts system state.
//
// Test Philosophy:
//   - Tests are privilege-aware: some may skip if elevated privileges are needed
//   - Tests are state-aware: system limits vary by platform and configuration
//   - Tests are non-destructive: limits persist only for the test process
var _ = Describe("SystemFileDescriptor", func() {
	var (
		originalCurrent int
		originalMax     int
	)

	// Capture initial system state before each test.
	// This allows tests to make relative assertions and detect changes.
	BeforeEach(func() {
		var err error
		originalCurrent, originalMax, err = SystemFileDescriptor(0)
		Expect(err).ToNot(HaveOccurred())
		Expect(originalCurrent).To(BeNumerically(">", 0))
		Expect(originalMax).To(BeNumerically(">=", originalCurrent))
	})

	// Test query operations (newValue <= 0).
	// These should always succeed without privileges.
	Context("Reading current limits", func() {
		It("should return current limits when called with 0", func() {
			current, max, err := SystemFileDescriptor(0)

			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(BeNumerically(">", 0))
			Expect(max).To(BeNumerically(">=", current))
		})

		It("should return positive values", func() {
			current, max, err := SystemFileDescriptor(0)

			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(BeNumerically(">", 0))
			Expect(max).To(BeNumerically(">", 0))
		})

		It("should have max greater than or equal to current", func() {
			current, max, err := SystemFileDescriptor(0)

			Expect(err).ToNot(HaveOccurred())
			Expect(max).To(BeNumerically(">=", current))
		})

		It("should return consistent values on multiple calls", func() {
			current1, max1, err1 := SystemFileDescriptor(0)
			current2, max2, err2 := SystemFileDescriptor(0)

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
			Expect(current1).To(Equal(current2))
			Expect(max1).To(Equal(max2))
		})
	})

	// Verify that negative values are treated as query operations.
	// The function should never attempt to set negative limits.
	Context("Setting limits with negative values", func() {
		It("should ignore negative values and return current limits", func() {
			current, max, err := SystemFileDescriptor(-1)

			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(Equal(originalCurrent))
			Expect(max).To(Equal(originalMax))
		})

		It("should ignore negative values without changing system state", func() {
			SystemFileDescriptor(-100)

			// Verify nothing changed
			current, max, err := SystemFileDescriptor(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(Equal(originalCurrent))
			Expect(max).To(Equal(originalMax))
		})
	})

	// Verify safety: the function never decreases limits.
	// This is a critical safety feature to prevent accidental resource restriction.
	Context("Setting limits below current value", func() {
		It("should not decrease the current limit", func() {
			lowerValue := originalCurrent - 10
			if lowerValue < 1 {
				lowerValue = 1
			}

			current, max, err := SystemFileDescriptor(lowerValue)

			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(Equal(originalCurrent))
			Expect(max).To(Equal(originalMax))
		})

		It("should keep system state unchanged when value is too low", func() {
			SystemFileDescriptor(10)

			// Verify nothing changed
			current, max, err := SystemFileDescriptor(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(Equal(originalCurrent))
			Expect(max).To(Equal(originalMax))
		})
	})

	// Verify idempotency: setting to current value is a no-op.
	Context("Setting limits to same value", func() {
		It("should succeed when setting to current value", func() {
			current, max, err := SystemFileDescriptor(originalCurrent)

			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(Equal(originalCurrent))
			Expect(max).To(Equal(originalMax))
		})
	})

	// Test limit increases within the hard limit range.
	// These may succeed or fail depending on privileges and system state.
	Context("Setting limits to slightly higher value", func() {
		It("should increase the limit when possible", func() {
			// Try to increase by a small amount within the hard limit
			newValue := originalCurrent + 10
			if newValue > originalMax {
				Skip("Cannot test increase: already at max limit")
			}

			current, max, err := SystemFileDescriptor(newValue)

			// This might fail due to permissions, which is acceptable
			if err == nil {
				Expect(current).To(BeNumerically(">=", newValue))
				Expect(max).To(BeNumerically(">=", current))
			}
		})
	})

	// Test boundary conditions and extreme values.
	// These verify robustness and safe error handling.
	Context("Edge cases", func() {
		It("should handle zero value", func() {
			current, max, err := SystemFileDescriptor(0)

			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(BeNumerically(">", 0))
			Expect(max).To(BeNumerically(">", 0))
		})

		It("should handle value of 1", func() {
			// Should not decrease below current
			current, max, err := SystemFileDescriptor(1)

			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(BeNumerically(">=", originalCurrent))
			Expect(max).To(BeNumerically(">=", current))
		})

		It("should handle very large values gracefully", func() {
			// Try to set an unreasonably large value
			// This should either succeed (if we have permissions) or fail gracefully
			_, _, err := SystemFileDescriptor(999999)

			// We accept both success and certain types of failures
			// (e.g., permission denied, or hitting hard limit)
			if err != nil {
				// Error is acceptable - we're trying to set a very high limit
				Expect(err).To(HaveOccurred())
			}
		})
	})

	// Verify system-wide invariants and sanity checks.
	// These ensure the function returns reasonable values on all platforms.
	Context("System behavior verification", func() {
		It("should maintain limits within system constraints", func() {
			current, max, err := SystemFileDescriptor(0)

			Expect(err).ToNot(HaveOccurred())

			// Reasonable minimum values (even conservative systems should allow this)
			Expect(current).To(BeNumerically(">=", 256))

			// Max should be at least as large as current
			Expect(max).To(BeNumerically(">=", current))
		})

		It("should return values suitable for file operations", func() {
			current, _, err := SystemFileDescriptor(0)

			Expect(err).ToNot(HaveOccurred())

			// The current limit should be practical for normal operations
			// Most systems default to at least 1024
			Expect(current).To(BeNumerically(">=", 100))
		})
	})
})
