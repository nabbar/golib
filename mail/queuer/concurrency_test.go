/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package queuer_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nabbar/golib/mail/queuer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Concurrency", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Concurrent Send Operations", func() {
		Context("with multiple goroutines", func() {
			It("should handle concurrent sends safely", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)

				numGoroutines := 10
				numMessagesPerRoutine := 5
				var wg sync.WaitGroup
				wg.Add(numGoroutines)

				errors := make([]error, 0)
				var errorsMux sync.Mutex

				for i := 0; i < numGoroutines; i++ {
					go func(routineID int) {
						defer wg.Done()
						for j := 0; j < numMessagesPerRoutine; j++ {
							err := pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d-%d", routineID, j)))
							if err != nil {
								errorsMux.Lock()
								errors = append(errors, err)
								errorsMux.Unlock()
							}
						}
					}(i)
				}

				wg.Wait()

				Expect(errors).To(BeEmpty())
				Expect(backend.msgCount.Load()).To(Equal(int32(numGoroutines * numMessagesPerRoutine)))
			})

			It("should maintain throttle limits with concurrent access", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  5,
					Wait: 200 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)

				numGoroutines := 3
				numMessages := 10
				var wg sync.WaitGroup
				wg.Add(numGoroutines)

				start := time.Now()

				for i := 0; i < numGoroutines; i++ {
					go func(routineID int) {
						defer wg.Done()
						for j := 0; j < numMessages; j++ {
							_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d-%d", routineID, j)))
						}
					}(i)
				}

				wg.Wait()
				elapsed := time.Since(start)

				// With 30 total messages and max=5, should require multiple throttle cycles
				expectedMinTime := 200 * time.Millisecond * time.Duration((numGoroutines*numMessages)/cfg.Max-1)
				Expect(elapsed).To(BeNumerically(">=", expectedMinTime))
				Expect(backend.msgCount.Load()).To(Equal(int32(numGoroutines * numMessages)))
			})
		})

		Context("with shared counter", func() {
			It("should not have race conditions", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  20,
					Wait: 50 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)

				numGoroutines := 20
				var wg sync.WaitGroup
				wg.Add(numGoroutines)

				for i := 0; i < numGoroutines; i++ {
					go func(routineID int) {
						defer wg.Done()
						_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d", routineID)))
					}(i)
				}

				wg.Wait()

				// All messages should be sent successfully
				Expect(backend.msgCount.Load()).To(Equal(int32(numGoroutines)))
			})
		})
	})

	Describe("Concurrent Clone Operations", func() {
		It("should safely clone from multiple goroutines", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)

			numGoroutines := 10
			var wg sync.WaitGroup
			wg.Add(numGoroutines)

			clones := make([]queuer.Pooler, numGoroutines)
			var clonesMux sync.Mutex

			for i := 0; i < numGoroutines; i++ {
				go func(idx int) {
					defer wg.Done()
					clone := pooler.NewPooler()
					clonesMux.Lock()
					clones[idx] = clone
					clonesMux.Unlock()
				}(i)
			}

			wg.Wait()

			// All clones should be valid and independent
			for i, clone := range clones {
				Expect(clone).ToNot(BeNil(), fmt.Sprintf("clone %d should not be nil", i))
				err := clone.Check(ctx)
				Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("clone %d should check successfully", i))
			}
		})

		It("should maintain independent state in clones", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  5,
				Wait: 200 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)

			// Create clones
			clone1 := pooler.NewPooler()
			clone2 := pooler.NewPooler()

			var wg sync.WaitGroup
			wg.Add(2)

			// Use up clone1's quota
			go func() {
				defer wg.Done()
				for i := 0; i < 5; i++ {
					_ = clone1.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("c1-msg-%d", i)))
				}
			}()

			// clone2 should still have full quota
			go func() {
				defer wg.Done()
				start := time.Now()
				for i := 0; i < 5; i++ {
					_ = clone2.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("c2-msg-%d", i)))
				}
				elapsed := time.Since(start)
				// Should be quick since clone2 has its own quota
				Expect(elapsed).To(BeNumerically("<", 150*time.Millisecond))
			}()

			wg.Wait()

			Expect(backend.msgCount.Load()).To(Equal(int32(10)))
		})
	})

	Describe("Concurrent Reset Operations", func() {
		It("should handle concurrent resets safely", func() {
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			callCount := atomic.Int32{}
			cfg.SetFuncCaller(func() error {
				callCount.Add(1)
				time.Sleep(10 * time.Millisecond) // Simulate work
				return nil
			})

			pooler := queuer.New(cfg, nil)

			numGoroutines := 10
			var wg sync.WaitGroup
			wg.Add(numGoroutines)

			errors := make([]error, 0)
			var errorsMux sync.Mutex

			for i := 0; i < numGoroutines; i++ {
				go func() {
					defer wg.Done()
					err := pooler.Reset()
					if err != nil {
						errorsMux.Lock()
						errors = append(errors, err)
						errorsMux.Unlock()
					}
				}()
			}

			wg.Wait()

			Expect(errors).To(BeEmpty())
			Expect(callCount.Load()).To(Equal(int32(numGoroutines)))
		})

		It("should handle concurrent reset and send", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  5,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)

			var wg sync.WaitGroup
			wg.Add(2)

			// Goroutine 1: Send messages
			go func() {
				defer wg.Done()
				for i := 0; i < 10; i++ {
					_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d", i)))
					time.Sleep(20 * time.Millisecond)
				}
			}()

			// Goroutine 2: Reset periodically
			go func() {
				defer wg.Done()
				for i := 0; i < 5; i++ {
					time.Sleep(50 * time.Millisecond)
					_ = pooler.Reset()
				}
			}()

			wg.Wait()

			// Should complete without deadlock or panic
			Expect(backend.msgCount.Load()).To(BeNumerically(">", 0))
		})
	})

	Describe("Concurrent Close Operations", func() {
		It("should handle concurrent close calls", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)

			numGoroutines := 10
			var wg sync.WaitGroup
			wg.Add(numGoroutines)

			for i := 0; i < numGoroutines; i++ {
				go func() {
					defer wg.Done()
					pooler.Close()
				}()
			}

			// Should not panic or deadlock
			Expect(func() { wg.Wait() }).ToNot(Panic())
		})

		It("should handle close during operations", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  5,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)

			var wg sync.WaitGroup
			wg.Add(2)

			// Goroutine 1: Send messages
			go func() {
				defer wg.Done()
				for i := 0; i < 20; i++ {
					_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d", i)))
					time.Sleep(10 * time.Millisecond)
				}
			}()

			// Goroutine 2: Close after a delay
			go func() {
				defer wg.Done()
				time.Sleep(100 * time.Millisecond)
				pooler.Close()
			}()

			// Should not panic
			Expect(func() { wg.Wait() }).ToNot(Panic())
		})
	})

	Describe("Stress Tests", func() {
		Context("with high concurrency", func() {
			It("should handle many concurrent operations", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  50,
					Wait: 50 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)

				numGoroutines := 100
				numMessagesPerRoutine := 10
				var wg sync.WaitGroup
				wg.Add(numGoroutines)

				successCount := atomic.Int32{}

				for i := 0; i < numGoroutines; i++ {
					go func(routineID int) {
						defer wg.Done()
						for j := 0; j < numMessagesPerRoutine; j++ {
							err := pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d-%d", routineID, j)))
							if err == nil {
								successCount.Add(1)
							}
						}
					}(i)
				}

				wg.Wait()

				// Should successfully send all messages
				Expect(successCount.Load()).To(Equal(int32(numGoroutines * numMessagesPerRoutine)))
			})

			It("should handle rapid clone and send", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  10,
					Wait: 50 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)

				numGoroutines := 50
				var wg sync.WaitGroup
				wg.Add(numGoroutines)

				successCount := atomic.Int32{}

				for i := 0; i < numGoroutines; i++ {
					go func(routineID int) {
						defer wg.Done()
						clone := pooler.NewPooler()
						err := clone.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d", routineID)))
						if err == nil {
							successCount.Add(1)
						}
					}(i)
				}

				wg.Wait()

				Expect(successCount.Load()).To(Equal(int32(numGoroutines)))
			})
		})

		Context("with mixed operations", func() {
			It("should handle mixed concurrent operations", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  10,
					Wait: 50 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)

				var wg sync.WaitGroup
				wg.Add(4)

				// Goroutine 1: Send messages
				go func() {
					defer wg.Done()
					for i := 0; i < 20; i++ {
						_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("send-%d", i)))
					}
				}()

				// Goroutine 2: Clone and send
				go func() {
					defer wg.Done()
					for i := 0; i < 10; i++ {
						clone := pooler.NewPooler()
						_ = clone.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("clone-%d", i)))
					}
				}()

				// Goroutine 3: Reset periodically
				go func() {
					defer wg.Done()
					for i := 0; i < 5; i++ {
						time.Sleep(30 * time.Millisecond)
						_ = pooler.Reset()
					}
				}()

				// Goroutine 4: Check periodically
				go func() {
					defer wg.Done()
					for i := 0; i < 10; i++ {
						time.Sleep(20 * time.Millisecond)
						_ = pooler.Check(ctx)
					}
				}()

				// Should complete without panic or deadlock
				Expect(func() { wg.Wait() }).ToNot(Panic())
				Expect(backend.msgCount.Load()).To(BeNumerically(">", 0))
			})
		})
	})

	Describe("Context Cancellation with Concurrency", func() {
		It("should handle context cancellation with concurrent sends", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  5,
				Wait: 200 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)

			localCtx, localCancel := context.WithCancel(ctx)

			numGoroutines := 10
			var wg sync.WaitGroup
			wg.Add(numGoroutines)

			errorCount := atomic.Int32{}

			for i := 0; i < numGoroutines; i++ {
				go func(routineID int) {
					defer wg.Done()
					for j := 0; j < 20; j++ {
						err := pooler.Send(localCtx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d-%d", routineID, j)))
						if err != nil {
							errorCount.Add(1)
							return
						}
					}
				}(i)
			}

			// Cancel context after a short delay
			time.Sleep(100 * time.Millisecond)
			localCancel()

			wg.Wait()

			// Some operations should have failed due to context cancellation
			Expect(errorCount.Load()).To(BeNumerically(">", 0))
		})
	})
})
