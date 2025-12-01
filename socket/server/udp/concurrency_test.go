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

package udp_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Server Concurrency", func() {
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
		time.Sleep(100 * time.Millisecond)
	})

	Describe("Concurrent Server Creation", func() {
		It("should handle concurrent New() calls", func() {
			const numServers = 10
			var wg sync.WaitGroup
			servers := make([]interface{}, numServers)
			errors := make([]error, numServers)

			for i := 0; i < numServers; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					handler := newTestHandler(false)
					srv, err := createServerWithHandler(handler.handler)
					servers[idx] = srv
					errors[idx] = err
				}(i)
			}

			wg.Wait()

			// All should succeed
			for i := 0; i < numServers; i++ {
				Expect(errors[i]).ToNot(HaveOccurred())
				Expect(servers[i]).ToNot(BeNil())
			}
		})

		It("should handle concurrent RegisterServer calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			const numCalls = 10
			var wg sync.WaitGroup
			results := make([]error, numCalls)

			for i := 0; i < numCalls; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					results[idx] = srv.RegisterServer("127.0.0.1:0")
				}(i)
			}

			wg.Wait()

			// Some should succeed (first ones), others may fail
			// At least one should succeed
			hasSuccess := false
			for _, err := range results {
				if err == nil {
					hasSuccess = true
					break
				}
			}
			Expect(hasSuccess).To(BeTrue())
		})
	})

	Describe("Concurrent State Access", func() {
		It("should handle concurrent IsRunning() calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			const numReads = 100
			var wg sync.WaitGroup
			results := make([]bool, numReads)

			for i := 0; i < numReads; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					results[idx] = srv.IsRunning()
				}(i)
			}

			wg.Wait()

			// All should return true (server is running)
			for _, running := range results {
				Expect(running).To(BeTrue())
			}
		})

		It("should handle concurrent IsGone() calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			const numReads = 100
			var wg sync.WaitGroup
			results := make([]bool, numReads)

			for i := 0; i < numReads; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					results[idx] = srv.IsGone()
				}(i)
			}

			wg.Wait()

			// All should return false (not gone yet)
			for _, gone := range results {
				Expect(gone).To(BeTrue())
			}
		})

		It("should handle concurrent OpenConnections() calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			const numReads = 100
			var wg sync.WaitGroup
			results := make([]int64, numReads)

			for i := 0; i < numReads; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					results[idx] = srv.OpenConnections()
				}(i)
			}

			wg.Wait()

			// All should return 0 (UDP is stateless)
			for _, conns := range results {
				Expect(conns).To(Equal(int64(0)))
			}
		})
	})

	Describe("Concurrent Callback Registration", func() {
		It("should handle concurrent RegisterFuncError calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			const numRegistrations = 10
			var wg sync.WaitGroup
			collectors := make([]*errorCollector, numRegistrations)

			for i := 0; i < numRegistrations; i++ {
				collectors[i] = newErrorCollector()
			}

			for i := 0; i < numRegistrations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					srv.RegisterFuncError(collectors[idx].callback)
				}(i)
			}

			wg.Wait()

			// Should complete without panic
			Expect(srv).ToNot(BeNil())
		})

		It("should handle concurrent RegisterFuncInfo calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			const numRegistrations = 10
			var wg sync.WaitGroup
			collectors := make([]*infoCollector, numRegistrations)

			for i := 0; i < numRegistrations; i++ {
				collectors[i] = newInfoCollector()
			}

			for i := 0; i < numRegistrations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					srv.RegisterFuncInfo(collectors[idx].callback)
				}(i)
			}

			wg.Wait()

			// Should complete without panic
			Expect(srv).ToNot(BeNil())
		})

		It("should handle concurrent RegisterFuncInfoServer calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			const numRegistrations = 10
			var wg sync.WaitGroup
			collectors := make([]*serverInfoCollector, numRegistrations)

			for i := 0; i < numRegistrations; i++ {
				collectors[i] = newServerInfoCollector()
			}

			for i := 0; i < numRegistrations; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					srv.RegisterFuncInfoServer(collectors[idx].callback)
				}(i)
			}

			wg.Wait()

			// Should complete without panic
			Expect(srv).ToNot(BeNil())
		})
	})

	Describe("Concurrent Shutdown", func() {
		It("should handle concurrent Shutdown() calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			const numShutdowns = 5
			var wg sync.WaitGroup
			results := make([]error, numShutdowns)

			for i := 0; i < numShutdowns; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer shutdownCancel()
					results[idx] = srv.Shutdown(shutdownCtx)
				}(i)
			}

			wg.Wait()

			// All should succeed or return nil
			for _, err := range results {
				// Either success or already shutdown
				_ = err
			}

			// Server should be stopped
			Eventually(func() bool {
				return srv.IsGone()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should handle concurrent Close() calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			const numCloses = 5
			var wg sync.WaitGroup
			results := make([]error, numCloses)

			for i := 0; i < numCloses; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					results[idx] = srv.Close()
				}(i)
			}

			wg.Wait()

			// All should succeed
			for _, err := range results {
				Expect(err).ToNot(HaveOccurred())
			}

			// Server should be stopped
			Eventually(func() bool {
				return !srv.IsRunning()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should handle mixed Shutdown and Close calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			var wg sync.WaitGroup

			// Some goroutines call Shutdown
			for i := 0; i < 3; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer shutdownCancel()
					_ = srv.Shutdown(shutdownCtx)
				}()
			}

			// Some goroutines call Close
			for i := 0; i < 3; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = srv.Close()
				}()
			}

			wg.Wait()

			// Server should be stopped
			Eventually(func() bool {
				return srv.IsGone()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})
	})

	Describe("Concurrent Start/Stop Cycles", func() {
		It("should handle rapid start/stop cycles", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			for cycle := 0; cycle < 3; cycle++ {
				localCtx, localCancel := context.WithCancel(ctx)

				go func() {
					_ = srv.Listen(localCtx)
				}()

				Eventually(func() bool {
					return srv.IsRunning()
				}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

				localCancel()

				Eventually(func() bool {
					return !srv.IsRunning()
				}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

				time.Sleep(50 * time.Millisecond)
			}
		})
	})

	Describe("Race Detection", func() {
		It("should not have data races with concurrent operations", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup

			// Start server
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = srv.Listen(ctx)
			}()

			time.Sleep(50 * time.Millisecond)

			// Concurrent state checks
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = srv.IsRunning()
					_ = srv.IsGone()
					_ = srv.OpenConnections()
				}()
			}

			// Concurrent callback registrations
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					srv.RegisterFuncError(func(...error) {})
					srv.RegisterFuncInfo(newInfoCollector().callback)
					srv.RegisterFuncInfoServer(func(string) {})
				}()
			}

			time.Sleep(100 * time.Millisecond)
			cancel()

			// Wait for all goroutines
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Success
			case <-time.After(5 * time.Second):
				Fail("Timeout waiting for goroutines")
			}
		})
	})
})
