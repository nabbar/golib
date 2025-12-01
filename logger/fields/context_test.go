/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package fields_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logfld "github.com/nabbar/golib/logger/fields"
)

var _ = Describe("Fields Context Integration", func() {
	Describe("Deadline", func() {
		Context("with deadline context", func() {
			It("should return deadline", func() {
				deadline := time.Now().Add(5 * time.Second)
				ctx, cancel := context.WithDeadline(context.Background(), deadline)
				defer cancel()

				f := logfld.New(ctx)

				dl, ok := f.Deadline()

				Expect(ok).To(BeTrue())
				Expect(dl).To(BeTemporally("~", deadline, 10*time.Millisecond))
			})
		})

		Context("without deadline", func() {
			It("should return false", func() {
				f := logfld.New(context.Background())

				_, ok := f.Deadline()

				Expect(ok).To(BeFalse())
			})
		})

		Context("with nil context", func() {
			It("should handle gracefully", func() {
				f := logfld.New(nil)

				_, ok := f.Deadline()

				Expect(ok).To(BeFalse())
			})
		})
	})

	Describe("Done", func() {
		Context("with cancelable context", func() {
			It("should close channel when context is canceled", func() {
				ctx, cancel := context.WithCancel(context.Background())
				f := logfld.New(ctx)

				done := f.Done()

				// Channel should not be closed initially
				select {
				case <-done:
					Fail("Done channel closed prematurely")
				default:
					// Expected
				}

				// Cancel context
				cancel()

				// Channel should close
				Eventually(done).Should(BeClosed())
			})

			It("should work with timeout", func() {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				f := logfld.New(ctx)

				done := f.Done()

				Eventually(done, 200*time.Millisecond).Should(BeClosed())
			})
		})

		Context("with background context", func() {
			It("should return never-closing channel", func() {
				f := logfld.New(context.Background())

				done := f.Done()

				select {
				case <-done:
					Fail("Background context Done should never close")
				case <-time.After(50 * time.Millisecond):
					// Expected
				}
			})
		})
	})

	Describe("Err", func() {
		Context("with canceled context", func() {
			It("should return Canceled error", func() {
				ctx, cancel := context.WithCancel(context.Background())
				f := logfld.New(ctx)

				cancel()

				// Give time for cancellation to propagate
				time.Sleep(10 * time.Millisecond)

				err := f.Err()

				Expect(err).To(Equal(context.Canceled))
			})
		})

		Context("with deadline exceeded", func() {
			It("should return DeadlineExceeded error", func() {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()

				f := logfld.New(ctx)

				// Wait for timeout
				time.Sleep(10 * time.Millisecond)

				err := f.Err()

				Expect(err).To(Equal(context.DeadlineExceeded))
			})
		})

		Context("with active context", func() {
			It("should return nil", func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				f := logfld.New(ctx)

				err := f.Err()

				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Value", func() {
		Context("with context values", func() {
			It("should retrieve context values", func() {
				type ctxKey string
				const key ctxKey = "test_key"

				ctx := context.WithValue(context.Background(), key, "test_value")
				f := logfld.New(ctx)

				val := f.Value(key)

				Expect(val).To(Equal("test_value"))
			})

			It("should return nil for non-existent keys", func() {
				f := logfld.New(context.Background())

				val := f.Value("non_existent")

				Expect(val).To(BeNil())
			})

			It("should preserve multiple context values", func() {
				type ctxKey string
				const key1 ctxKey = "key1"
				const key2 ctxKey = "key2"

				ctx := context.WithValue(context.Background(), key1, "value1")
				ctx = context.WithValue(ctx, key2, "value2")

				f := logfld.New(ctx)

				Expect(f.Value(key1)).To(Equal("value1"))
				Expect(f.Value(key2)).To(Equal("value2"))
			})
		})
	})

	Describe("GetContext", func() {
		Context("with valid context", func() {
			It("should return underlying context", func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				f := logfld.New(ctx)

				Expect(f).ToNot(BeNil())
				// Should have same cancellation behavior
				cancel()
				time.Sleep(10 * time.Millisecond)
				Expect(f.Err()).To(Equal(context.Canceled))
			})

			It("should preserve context values", func() {
				type ctxKey string
				const key ctxKey = "test"

				ctx := context.WithValue(context.Background(), key, "value")
				f := logfld.New(ctx)

				Expect(f.Value(key)).To(Equal("value"))
			})
		})
	})

	Describe("Context Interface Compatibility", func() {
		It("should implement context.Context interface", func() {
			f := logfld.New(context.Background())

			var _ context.Context = f
		})

		It("should work with context-aware functions", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			f := logfld.New(ctx)
			f.Add("test", "value")

			// Use as context.Context
			testFunc := func(ctx context.Context) bool {
				select {
				case <-ctx.Done():
					return false
				default:
					return true
				}
			}

			Expect(testFunc(f)).To(BeTrue())

			cancel()
			time.Sleep(10 * time.Millisecond)

			Expect(testFunc(f)).To(BeFalse())
		})
	})

	Describe("Context Propagation", func() {
		Context("with parent context", func() {
			It("should respect parent cancellation", func() {
				parent, parentCancel := context.WithCancel(context.Background())
				defer parentCancel()

				child, childCancel := context.WithCancel(parent)
				defer childCancel()

				f := logfld.New(child)

				// Cancel parent should affect fields
				parentCancel()
				time.Sleep(10 * time.Millisecond)

				Expect(f.Err()).ToNot(BeNil())
			})

			It("should respect parent deadline", func() {
				parent, parentCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
				defer parentCancel()

				f := logfld.New(parent)

				// Wait for parent timeout
				time.Sleep(100 * time.Millisecond)

				Expect(f.Err()).To(Equal(context.DeadlineExceeded))
			})
		})
	})

	Describe("Concurrent Context Access", func() {
		It("should be safe for concurrent reads", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			f := logfld.New(ctx)
			f.Add("key", "value")

			done := make(chan bool)

			// Multiple goroutines reading context methods
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					for j := 0; j < 100; j++ {
						_ = f.Err()
						_, _ = f.Deadline()
						_ = f.Done()
						_ = f.Value("key")
					}
					done <- true
				}()
			}

			// Wait for all goroutines
			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})
})
