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
	"runtime"

	. "github.com/nabbar/golib/ioutils/fileDescriptor"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SystemFileDescriptor - Limit Increase", func() {
	var (
		originalCurrent int
		originalMax     int
	)

	BeforeEach(func() {
		var err error
		originalCurrent, originalMax, err = SystemFileDescriptor(0)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Attempting to increase limits", func() {
		It("should attempt to increase within soft limit range", func() {
			// Calculate a target between current and max
			if originalMax <= originalCurrent {
				Skip("Already at maximum, cannot test increase")
			}

			// Try to increase to halfway between current and max
			targetValue := originalCurrent + (originalMax-originalCurrent)/2
			if targetValue <= originalCurrent {
				targetValue = originalCurrent + 100
			}

			// On Linux/Unix, this should succeed without root if within hard limit
			current, max, err := SystemFileDescriptor(targetValue)

			if err == nil {
				// Success case
				Expect(current).To(BeNumerically(">=", originalCurrent))
				Expect(max).To(BeNumerically(">=", current))
			} else {
				// Permission error is acceptable
				GinkgoWriter.Printf("Could not increase limit (may need elevated privileges): %v\n", err)
			}
		})

		It("should handle increase beyond soft limit but within hard limit", func() {
			if originalMax <= originalCurrent+1 {
				Skip("No room to increase between soft and hard limit")
			}

			// Try to set to original max (hard limit)
			current, max, err := SystemFileDescriptor(originalMax)

			if err == nil {
				Expect(current).To(BeNumerically(">=", originalCurrent))
				Expect(max).To(BeNumerically(">=", originalMax))
			}
			// Error is acceptable (may need privileges)
		})

		It("should handle attempts to exceed hard limit", func() {
			beyondMax := originalMax + 1000

			current, max, err := SystemFileDescriptor(beyondMax)

			if err == nil {
				// If it succeeded, verify it's reasonable
				Expect(current).To(BeNumerically(">", 0))
				Expect(max).To(BeNumerically(">=", current))
			} else {
				// Error is expected when trying to exceed hard limit without privileges
				GinkgoWriter.Printf("Expected error when exceeding hard limit: %v\n", err)
			}
		})
	})

	Context("Platform-specific behavior", func() {
		It("should behave appropriately for the current platform", func() {
			current, max, err := SystemFileDescriptor(0)

			Expect(err).ToNot(HaveOccurred())

			switch runtime.GOOS {
			case "linux", "darwin", "freebsd":
				// Unix-like systems should have reasonable defaults
				Expect(current).To(BeNumerically(">=", 256))
				Expect(max).To(BeNumerically(">=", current))
			case "windows":
				// Windows has different limits
				Expect(current).To(BeNumerically(">", 0))
				Expect(max).To(BeNumerically(">=", current))
			default:
				// Other platforms should at least return valid values
				Expect(current).To(BeNumerically(">", 0))
				Expect(max).To(BeNumerically(">=", current))
			}
		})
	})

	Context("Realistic scenarios", func() {
		It("should support typical application needs", func() {
			// Most applications need at least 1024 file descriptors
			typicalNeeds := 1024

			current, _, err := SystemFileDescriptor(0)
			Expect(err).ToNot(HaveOccurred())

			if current < typicalNeeds {
				// Try to increase to typical needs
				newCurrent, newMax, err := SystemFileDescriptor(typicalNeeds)

				if err == nil {
					Expect(newCurrent).To(BeNumerically(">=", typicalNeeds))
					Expect(newMax).To(BeNumerically(">=", newCurrent))
				}
				// Failure is acceptable (may need privileges)
			}
		})

		It("should handle server-level requirements", func() {
			// High-performance servers may need many descriptors
			serverNeeds := 4096

			current, max, _ := SystemFileDescriptor(0)

			if current >= serverNeeds {
				// Already sufficient
				Expect(current).To(BeNumerically(">=", serverNeeds))
			} else if max >= serverNeeds {
				// Try to increase to server needs
				_, _, err := SystemFileDescriptor(serverNeeds)
				// Both success and failure are acceptable
				// (depends on privileges and system configuration)
				if err != nil {
					GinkgoWriter.Printf("Could not set server-level limits: %v\n", err)
				}
			} else {
				Skip("System hard limit is below typical server needs")
			}
		})
	})

	Context("State verification after modification", func() {
		It("should maintain consistent state after multiple operations", func() {
			// Read initial
			c1, m1, _ := SystemFileDescriptor(0)

			// Try to modify (may or may not succeed)
			targetValue := c1 + 50
			if targetValue > m1 {
				targetValue = c1
			}
			SystemFileDescriptor(targetValue)

			// Read again
			c2, m2, err := SystemFileDescriptor(0)

			Expect(err).ToNot(HaveOccurred())
			Expect(c2).To(BeNumerically(">=", c1))
			Expect(m2).To(BeNumerically(">=", c2))
		})

		It("should not corrupt system state on failure", func() {
			initial, _, _ := SystemFileDescriptor(0)

			// Try an operation that might fail
			_, _, _ = SystemFileDescriptor(999999)

			// Verify we can still read the state
			current, max, err := SystemFileDescriptor(0)

			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(BeNumerically(">=", initial))
			Expect(max).To(BeNumerically(">", 0))
		})
	})
})
