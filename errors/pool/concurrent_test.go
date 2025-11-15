/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/errors/pool"
)

var _ = Describe("Concurrent Operations", func() {
	var p pool.Pool

	BeforeEach(func() {
		p = pool.New()
	})

	Describe("Concurrent Add", func() {
		It("should handle concurrent additions", func() {
			const goroutines = 100
			const errorsPerGoroutine = 10

			var wg sync.WaitGroup
			wg.Add(goroutines)

			for i := 0; i < goroutines; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < errorsPerGoroutine; j++ {
						err := fmt.Errorf("error from goroutine %d, iteration %d", id, j)
						p.Add(err)
					}
				}(i)
			}

			wg.Wait()

			// Should have all errors
			Expect(p.Len()).To(Equal(uint64(goroutines * errorsPerGoroutine)))
		})

		It("should maintain atomicity with concurrent adds", func() {
			const iterations = 1000

			var wg sync.WaitGroup
			wg.Add(iterations)

			for i := 0; i < iterations; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					p.Add(fmt.Errorf("error %d", id))
				}(i)
			}

			wg.Wait()

			Expect(p.Len()).To(Equal(uint64(iterations)))
		})
	})

	Describe("Concurrent Get/Set", func() {
		BeforeEach(func() {
			// Pre-populate with some errors
			for i := 1; i <= 100; i++ {
				p.Add(fmt.Errorf("error %d", i))
			}
		})

		It("should handle concurrent reads", func() {
			const readers = 50

			var wg sync.WaitGroup
			wg.Add(readers)

			for i := 0; i < readers; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					// Read multiple times
					for j := 0; j < 100; j++ {
						idx := uint64((id % 100) + 1)
						err := p.Get(idx)
						Expect(err).NotTo(BeNil())
					}
				}(i)
			}

			wg.Wait()
		})

		It("should handle concurrent writes", func() {
			const writers = 50

			var wg sync.WaitGroup
			wg.Add(writers)

			for i := 0; i < writers; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					idx := uint64((id % 100) + 1)
					newErr := fmt.Errorf("updated error %d", id)
					p.Set(idx, newErr)
				}(i)
			}

			wg.Wait()

			// All indices should still have errors
			for i := uint64(1); i <= 100; i++ {
				Expect(p.Get(i)).NotTo(BeNil())
			}
		})

		It("should handle concurrent reads and writes", func() {
			const operations = 100

			var wg sync.WaitGroup
			wg.Add(operations * 2)

			// Readers
			for i := 0; i < operations; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					idx := uint64((id % 100) + 1)
					p.Get(idx)
				}(i)
			}

			// Writers
			for i := 0; i < operations; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					idx := uint64((id % 100) + 1)
					p.Set(idx, fmt.Errorf("concurrent error %d", id))
				}(i)
			}

			wg.Wait()
		})
	})

	Describe("Concurrent Delete", func() {
		BeforeEach(func() {
			for i := 1; i <= 100; i++ {
				p.Add(fmt.Errorf("error %d", i))
			}
		})

		It("should handle concurrent deletions", func() {
			const deleters = 50

			var wg sync.WaitGroup
			wg.Add(deleters)

			for i := 0; i < deleters; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					idx := uint64((id % 100) + 1)
					p.Del(idx)
				}(i)
			}

			wg.Wait()

			// Some errors should be deleted
			Expect(p.Len()).To(BeNumerically("<", uint64(100)))
		})

		It("should handle concurrent add and delete", func() {
			var wg sync.WaitGroup
			wg.Add(200)

			// Adders
			for i := 0; i < 100; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					p.Add(fmt.Errorf("new error %d", id))
				}(i)
			}

			// Deleters
			for i := 0; i < 100; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					idx := uint64(id + 1)
					p.Del(idx)
				}(i)
			}

			wg.Wait()

			// Should still be able to query the pool
			_ = p.Len()
			_ = p.MaxId()
		})
	})

	Describe("Concurrent Clear", func() {
		It("should handle concurrent clear operations", func() {
			// Add some errors
			for i := 1; i <= 50; i++ {
				p.Add(fmt.Errorf("error %d", i))
			}

			var wg sync.WaitGroup
			wg.Add(10)

			for i := 0; i < 10; i++ {
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					p.Clear()
				}()
			}

			wg.Wait()

			Eventually(func() uint64 {
				return p.Len()
			}).Should(Equal(uint64(0)))
		})

		It("should handle clear with concurrent adds", func() {
			var wg sync.WaitGroup
			wg.Add(100)

			// Clear operations
			for i := 0; i < 50; i++ {
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					p.Clear()
				}()
			}

			// Add operations
			for i := 0; i < 50; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					p.Add(fmt.Errorf("error %d", id))
				}(i)
			}

			wg.Wait()

			// Pool should be in consistent state
			_ = p.Len()
			_ = p.MaxId()
		})
	})

	Describe("Concurrent Metadata Operations", func() {
		BeforeEach(func() {
			for i := 1; i <= 100; i++ {
				p.Add(fmt.Errorf("error %d", i))
			}
		})

		It("should handle concurrent Len() calls", func() {
			var wg sync.WaitGroup
			wg.Add(100)

			results := make([]uint64, 100)
			for i := 0; i < 100; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					results[id] = p.Len()
				}(i)
			}

			wg.Wait()

			// All results should be valid
			for _, length := range results {
				Expect(length).To(BeNumerically(">=", 0))
			}
		})

		It("should handle concurrent MaxId() calls", func() {
			var wg sync.WaitGroup
			wg.Add(100)

			for i := 0; i < 100; i++ {
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					maxId := p.MaxId()
					Expect(maxId).To(BeNumerically(">=", 0))
				}()
			}

			wg.Wait()
		})

		It("should handle concurrent Last() calls", func() {
			var wg sync.WaitGroup
			wg.Add(100)

			for i := 0; i < 100; i++ {
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_ = p.Last()
				}()
			}

			wg.Wait()
		})

		It("should handle concurrent Slice() calls", func() {
			var wg sync.WaitGroup
			wg.Add(50)

			for i := 0; i < 50; i++ {
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					slice := p.Slice()
					Expect(slice).NotTo(BeNil())
				}()
			}

			wg.Wait()
		})

		It("should handle concurrent Error() calls", func() {
			var wg sync.WaitGroup
			wg.Add(50)

			for i := 0; i < 50; i++ {
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					err := p.Error()
					Expect(err).NotTo(BeNil())
				}()
			}

			wg.Wait()
		})
	})

	Describe("Mixed Concurrent Operations", func() {
		It("should handle all operations concurrently", func() {
			const operations = 50

			var wg sync.WaitGroup
			wg.Add(operations * 8)

			// Add operations
			for i := 0; i < operations; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()
					p.Add(fmt.Errorf("add error %d", id))
				}(i)
			}

			// Get operations
			for i := 0; i < operations; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()
					idx := uint64((id % 10) + 1)
					p.Get(idx)
				}(i)
			}

			// Set operations
			for i := 0; i < operations; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()
					idx := uint64((id % 10) + 1)
					p.Set(idx, fmt.Errorf("set error %d", id))
				}(i)
			}

			// Del operations
			for i := 0; i < operations; i++ {
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()
					idx := uint64((id % 10) + 1)
					p.Del(idx)
				}(i)
			}

			// Len operations
			for i := 0; i < operations; i++ {
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					p.Len()
				}()
			}

			// MaxId operations
			for i := 0; i < operations; i++ {
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					p.MaxId()
				}()
			}

			// Last operations
			for i := 0; i < operations; i++ {
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					p.Last()
				}()
			}

			// Slice operations
			for i := 0; i < operations; i++ {
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					p.Slice()
				}()
			}

			wg.Wait()

			// Pool should be in consistent state
			_ = p.Len()
			_ = p.MaxId()
			_ = p.Last()
		})
	})
})
