//go:build linux || darwin

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

package unixgram_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unix Datagram Server Concurrency", func() {
	var (
		ctx      context.Context
		cancel   context.CancelFunc
		sockPath string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
		cleanupSocketFile(sockPath)
		time.Sleep(50 * time.Millisecond)
	})

	Describe("Concurrent Datagram Sending", func() {
		It("should handle concurrent senders", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			// Send datagrams from multiple goroutines
			var wg sync.WaitGroup
			numSenders := 10
			datagramsPerSender := 5

			for i := 0; i < numSenders; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < datagramsPerSender; j++ {
						_ = sendUnixgramDatagram(sockPath, []byte("message"))
						time.Sleep(time.Millisecond)
					}
				}(i)
			}

			wg.Wait()

			// Wait for all datagrams to be received
			Eventually(func() int64 {
				return handler.getCount()
			}, 5*time.Second, 10*time.Millisecond).Should(BeNumerically(">=", int64(numSenders*datagramsPerSender)))
		})

		It("should handle rapid concurrent sends", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			// Rapid concurrent sends
			var wg sync.WaitGroup
			numSenders := 20
			datagramsPerSender := 10

			for i := 0; i < numSenders; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < datagramsPerSender; j++ {
						_ = sendUnixgramDatagram(sockPath, []byte("fast"))
					}
				}()
			}

			wg.Wait()

			// Should receive most datagrams
			Eventually(func() int64 {
				return handler.getCount()
			}, 5*time.Second, 10*time.Millisecond).Should(BeNumerically(">", int64(numSenders*datagramsPerSender)/2))
		})
	})

	Describe("Concurrent State Queries", func() {
		It("should handle concurrent IsRunning calls", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			// Multiple goroutines checking state
			var wg sync.WaitGroup
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						_ = srv.IsRunning()
						_ = srv.IsGone()
						_ = srv.OpenConnections()
					}
				}()
			}

			wg.Wait()
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Describe("Concurrent Callback Registration", func() {
		It("should handle concurrent callback registration", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					collector := newInfoCollector()
					srv.RegisterFuncInfo(collector.callback)
				}()
			}

			wg.Wait()
			startServer(srv, ctx)

			// Server should still work
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Describe("Concurrent Shutdown", func() {
		It("should handle concurrent shutdown calls", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			// Multiple goroutines trying to shutdown
			var wg sync.WaitGroup
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer shutdownCancel()
					_ = srv.Shutdown(shutdownCtx)
				}()
			}

			wg.Wait()

			Eventually(func() bool {
				return srv.IsGone()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})
	})
})
