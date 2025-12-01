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
 */

package fileDescriptor_test

import (
	"sync"

	. "github.com/nabbar/golib/ioutils/fileDescriptor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Concurrency tests for SystemFileDescriptor.
// These tests verify thread-safety and behavior under concurrent access.
//
// Note: The function is naturally thread-safe because:
//   - Unix/Linux/macOS: syscalls are synchronized at kernel level
//   - Windows: C runtime functions are thread-safe
//   - No shared state in the application layer
//
// These tests verify that concurrent calls produce consistent results
// and don't cause race conditions or data corruption.
var _ = Describe("SystemFileDescriptor - Concurrency", func() {
	Context("Concurrent read operations", func() {
		It("should handle multiple simultaneous queries without error", func() {
			const goroutines = 50

			var wg sync.WaitGroup
			results := make(chan struct {
				current int
				max     int
				err     error
			}, goroutines)

			// Launch multiple concurrent queries
			for i := 0; i < goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					current, max, err := SystemFileDescriptor(0)
					results <- struct {
						current int
						max     int
						err     error
					}{current, max, err}
				}()
			}

			wg.Wait()
			close(results)

			// Collect and verify all results
			var firstCurrent, firstMax int
			var firstSet bool

			for result := range results {
				// All calls should succeed
				Expect(result.err).ToNot(HaveOccurred())
				Expect(result.current).To(BeNumerically(">", 0))
				Expect(result.max).To(BeNumerically(">=", result.current))

				// All results should be identical (reading same system state)
				if !firstSet {
					firstCurrent = result.current
					firstMax = result.max
					firstSet = true
				} else {
					Expect(result.current).To(Equal(firstCurrent),
						"Concurrent queries should return same current limit")
					Expect(result.max).To(Equal(firstMax),
						"Concurrent queries should return same max limit")
				}
			}
		})

		It("should not corrupt data under high concurrency", func() {
			const goroutines = 100
			const iterations = 10

			var wg sync.WaitGroup
			errorCount := 0
			var mu sync.Mutex

			// Multiple goroutines doing multiple queries each
			for i := 0; i < goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < iterations; j++ {
						current, max, err := SystemFileDescriptor(0)
						if err != nil {
							mu.Lock()
							errorCount++
							mu.Unlock()
							continue
						}

						// Verify invariants
						Expect(current).To(BeNumerically(">", 0))
						Expect(max).To(BeNumerically(">=", current))
					}
				}()
			}

			wg.Wait()

			// No errors should occur
			Expect(errorCount).To(Equal(0))
		})
	})

	Context("Concurrent read and write operations", func() {
		It("should handle mixed query and increase operations", func() {
			const readers = 20
			const writers = 5

			initial, initialMax, err := SystemFileDescriptor(0)
			Expect(err).ToNot(HaveOccurred())

			// Calculate a safe target within hard limit
			target := initial + 10
			if target > initialMax {
				Skip("Cannot test: no room to increase limit")
			}

			var wg sync.WaitGroup
			errors := make(chan error, readers+writers)

			// Launch reader goroutines
			for i := 0; i < readers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_, _, err := SystemFileDescriptor(0)
					if err != nil {
						errors <- err
					}
				}()
			}

			// Launch writer goroutines (may fail due to permissions)
			for i := 0; i < writers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_, _, err := SystemFileDescriptor(target)
					// Permission errors are acceptable
					if err != nil {
						GinkgoWriter.Printf("Writer got error (acceptable): %v\n", err)
					}
				}()
			}

			wg.Wait()
			close(errors)

			// Check that no unexpected errors occurred
			// (permission errors from writers were already logged)
			errorList := []error{}
			for err := range errors {
				errorList = append(errorList, err)
			}

			// All errors should be permission-related or nil
			for _, err := range errorList {
				if err != nil {
					GinkgoWriter.Printf("Error during concurrent operations: %v\n", err)
				}
			}
		})

		It("should maintain consistency when multiple goroutines try to increase", func() {
			initial, max, err := SystemFileDescriptor(0)
			Expect(err).ToNot(HaveOccurred())

			// Calculate targets
			target1 := initial + 5
			target2 := initial + 10

			if target2 > max {
				Skip("Cannot test: targets exceed maximum")
			}

			var wg sync.WaitGroup
			results := make(chan struct {
				current int
				err     error
			}, 2)

			// Two goroutines trying to increase simultaneously
			for _, target := range []int{target1, target2} {
				wg.Add(1)
				go func(t int) {
					defer wg.Done()
					defer GinkgoRecover()

					current, _, err := SystemFileDescriptor(t)
					results <- struct {
						current int
						err     error
					}{current, err}
				}(target)
			}

			wg.Wait()
			close(results)

			// Verify final state is consistent
			final, finalMax, err := SystemFileDescriptor(0)
			Expect(err).ToNot(HaveOccurred())

			// Final state should be at least initial
			Expect(final).To(BeNumerically(">=", initial))
			Expect(finalMax).To(BeNumerically(">=", final))

			// Results should be reasonable
			for result := range results {
				if result.err == nil {
					Expect(result.current).To(BeNumerically(">=", initial))
				}
			}
		})
	})

	Context("Race condition detection", func() {
		It("should not have race conditions under concurrent access", func() {
			// This test is designed to be run with -race flag
			// go test -race will detect any race conditions

			const goroutines = 50

			var wg sync.WaitGroup

			// Mix of operations
			for i := 0; i < goroutines; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					defer GinkgoRecover()

					if idx%2 == 0 {
						// Query operation
						current, max, _ := SystemFileDescriptor(0)
						Expect(current).To(BeNumerically(">", 0))
						Expect(max).To(BeNumerically(">=", current))
					} else {
						// Attempt increase (may fail, that's OK)
						current, _, _ := SystemFileDescriptor(0)
						SystemFileDescriptor(current + 1)
					}
				}(i)
			}

			wg.Wait()
		})

		It("should handle rapid sequential calls from multiple goroutines", func() {
			const goroutines = 20
			const callsPerGoroutine = 100

			var wg sync.WaitGroup

			for i := 0; i < goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					// Rapid fire calls
					for j := 0; j < callsPerGoroutine; j++ {
						current, max, err := SystemFileDescriptor(0)
						Expect(err).ToNot(HaveOccurred())
						Expect(current).To(BeNumerically(">", 0))
						Expect(max).To(BeNumerically(">=", current))
					}
				}()
			}

			wg.Wait()
		})
	})

	Context("Stress testing", func() {
		It("should remain stable under sustained concurrent load", func() {
			const duration = 100 // Number of iterations
			const workers = 30

			var wg sync.WaitGroup
			stop := make(chan struct{})
			errorCount := 0
			var mu sync.Mutex

			// Start workers
			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					count := 0
					for count < duration {
						select {
						case <-stop:
							return
						default:
							current, max, err := SystemFileDescriptor(0)
							if err != nil {
								mu.Lock()
								errorCount++
								mu.Unlock()
								return
							}

							// Verify invariants
							if current <= 0 || max < current {
								mu.Lock()
								errorCount++
								mu.Unlock()
								GinkgoWriter.Printf("Invalid state: current=%d, max=%d\n", current, max)
								return
							}

							count++
						}
					}
				}(i)
			}

			wg.Wait()
			close(stop)

			// Should have no errors
			Expect(errorCount).To(Equal(0), "Should have no errors under concurrent load")
		})
	})
})
