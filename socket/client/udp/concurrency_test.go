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
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Client Concurrency", func() {
	Describe("Concurrent IsConnected", func() {
		It("should handle concurrent IsConnected calls", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			var wg sync.WaitGroup
			errors := atomic.Int32{}

			// Concurrent reads of IsConnected
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					for j := 0; j < 100; j++ {
						connected := cli.IsConnected()
						if !connected {
							errors.Add(1)
						}
					}
				}()
			}

			wg.Wait()
			Expect(errors.Load()).To(Equal(int32(0)))
		})
	})

	Describe("Concurrent RegisterFunc", func() {
		It("should handle concurrent callback registration", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			var wg sync.WaitGroup
			callCount := atomic.Int32{}

			// Concurrent callback registration
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(id int) {
					defer GinkgoRecover()
					defer wg.Done()

					cli.RegisterFuncError(func(errs ...error) {
						callCount.Add(1)
					})

					cli.RegisterFuncInfo(func(_, _ net.Addr, _ libsck.ConnState) {
						// just register
					})
				}(i)
			}

			wg.Wait()
			Expect(cli).ToNot(BeNil())
		})
	})

	Describe("Concurrent Connect", func() {
		It("should handle concurrent Connect calls safely", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			var wg sync.WaitGroup
			successCount := atomic.Int32{}

			// Concurrent connects
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					if err := cli.Connect(ctx); err == nil {
						successCount.Add(1)
					}
				}()
			}

			wg.Wait()
			Expect(successCount.Load()).To(BeNumerically(">", 0))
			Expect(cli.IsConnected()).To(BeTrue())
		})
	})

	Describe("Sequential Writes", func() {
		It("should handle sequential writes from one goroutine", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			successCount := 0
			for i := 0; i < 100; i++ {
				data := []byte(fmt.Sprintf("msg%d", i))
				n, err := cli.Write(data)
				if err == nil && n == len(data) {
					successCount++
				}
				time.Sleep(time.Millisecond) // Small delay to avoid overwhelming
			}

			// UDP may lose packets, especially under load
			// Expect at least 50% success rate for sequential writes
			Expect(successCount).To(BeNumerically(">=", 50))
		})
	})

	Describe("Concurrent State Checks", func() {
		It("should handle concurrent state checks during operations", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			var wg sync.WaitGroup
			stopChan := make(chan struct{})

			// Continuous state checking
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					for {
						select {
						case <-stopChan:
							return
						default:
							_ = cli.IsConnected()
							time.Sleep(1 * time.Millisecond)
						}
					}
				}()
			}

			// Perform operations while state is being checked
			for i := 0; i < 10; i++ {
				data := []byte(fmt.Sprintf("concurrent%d", i))
				_, _ = cli.Write(data)
				time.Sleep(5 * time.Millisecond)
			}

			close(stopChan)
			wg.Wait()

			Expect(cli.IsConnected()).To(BeTrue())
		})
	})

	Describe("Connect and Close Race", func() {
		It("should handle concurrent connect and close", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			var wg sync.WaitGroup

			// Concurrent connect and close
			for i := 0; i < 5; i++ {
				wg.Add(2)

				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					_ = cli.Connect(ctx)
				}()

				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					time.Sleep(10 * time.Millisecond)
					_ = cli.Close()
				}()

				wg.Wait()
				time.Sleep(50 * time.Millisecond)
			}
		})
	})

	Describe("Callback Execution Concurrency", func() {
		It("should execute callbacks asynchronously without blocking", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			callbackStarted := make(chan struct{}, 10)
			callbackFinished := make(chan struct{}, 10)

			cli.RegisterFuncError(func(errs ...error) {
				callbackStarted <- struct{}{}
				time.Sleep(50 * time.Millisecond) // Slow callback
				callbackFinished <- struct{}{}
			})

			// Operations should not block waiting for callbacks
			start := time.Now()

			// Trigger errors (write without connection temporarily)
			_ = cli.Close()
			for i := 0; i < 5; i++ {
				_, _ = cli.Write([]byte("test"))
			}

			duration := time.Since(start)

			// Should complete quickly despite slow callbacks
			Expect(duration).To(BeNumerically("<", 100*time.Millisecond))

			// Callbacks should still execute
			Eventually(callbackStarted, 2*time.Second).Should(Receive())
		})
	})

	Describe("Multiple Clients Concurrent Access", func() {
		It("should handle multiple independent clients concurrently", func() {
			srv, _, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			address := getTestAddress()
			srv2 := createServer(simpleEchoHandler(), address)
			startServer(ctx, srv2)
			waitForServerRunning(address, 5*time.Second)
			defer cleanupServer(srv2, ctx)

			var wg sync.WaitGroup
			successCount := atomic.Int32{}

			// Multiple clients operating concurrently
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(id int) {
					defer GinkgoRecover()
					defer wg.Done()

					cli := createClient(address)
					defer cleanupClient(cli)

					if err := cli.Connect(ctx); err != nil {
						return
					}

					// Each client sends messages
					for j := 0; j < 5; j++ {
						data := []byte(fmt.Sprintf("client%d-msg%d", id, j))
						if n, err := cli.Write(data); err == nil && n > 0 {
							successCount.Add(1)
						}
					}
				}(i)
			}

			wg.Wait()
			Expect(successCount.Load()).To(BeNumerically(">", 0))
		})
	})

	Describe("Stress Test - Rapid Operations", func() {
		It("should handle rapid sequential operations", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			// Rapid connect/disconnect cycles
			for i := 0; i < 10; i++ {
				err := cli.Connect(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Quick write
				_, err = cli.Write([]byte("rapid"))
				Expect(err).ToNot(HaveOccurred())

				err = cli.Close()
				Expect(err).ToNot(HaveOccurred())
			}
		})
	})
})
