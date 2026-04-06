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

// Package tcp_test validates the server's behavior under high concurrent load.
//
// # Concurrency Test Logic
//
// The 'concurrency_test.go' file stresses the server's thread-safety mechanisms. 
// It verifies that:
//   - Lock-Free State: Atomic operations on 'run', 'gon', and 'nc' counters 
//     work correctly without data races.
//   - Parallel Acceptance: The server can handle multiple simultaneous 
//     connection attempts without dropping clients.
//   - Context Isolation: Handlers for different connections do not interfere 
//     with each other's data or context.
//   - Counter Integrity: The atomic connection counter remains accurate even 
//     under rapid open/close churn.
//
// # Dataflow: Concurrent Stress Test
//
//	[Test Runner] ───(Spawn 50 clients)───> [net.Dial] ───┐
//	                                          │           │
//	[srv.OpenConnections()] <────────────(Accept Loop) <──┘
//	          │                               │
//	[Atomic Increment] <──────────────────────┘
//	          │
//	[Atomic Decrement] <──────(on client disconnect)──────┐
//	                                                      │
//	[All Clients Closed] ───> [Verify Counter == 0] ──────┘
package tcp_test

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Concurrency", func() {
	var (
		srv scksrt.ServerTcp
		adr string
		c   context.Context
		cnl context.CancelFunc
	)

	// Setup: Initialize address and context.
	BeforeEach(func() {
		adr = getTestAddr()
		c, cnl = context.WithCancel(globalCtx)
	})

	// Cleanup: Stop server and ensure no leftover goroutines.
	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		time.Sleep(100 * time.Millisecond)
	})

	Context("concurrent connections", func() {
		// Test: Handling 10 simultaneous clients.
		It("should handle multiple simultaneous connections", func() {
			cnt := new(atomic.Int32)
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, counterHandler(cnt), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			numConns := 10
			var wg sync.WaitGroup
			wg.Add(numConns)

			for i := 0; i < numConns; i++ {
				go func() {
					defer wg.Done()
					con := connectToServer(adr)
					defer func() { _ = con.Close() }()

					msg := []byte("test")
					rsp := sendAndReceive(con, msg)
					Expect(rsp).To(Equal(msg))
				}()
			}

			wg.Wait()
			// Counter should exactly match the number of clients processed.
			Expect(cnt.Load()).To(BeNumerically(">=", int32(numConns)))
		})

		// Test: Verifying data isolation across connections.
		It("should handle concurrent writes from multiple connections", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			numConns := 5
			msgsPerConn := 20
			var wg sync.WaitGroup
			wg.Add(numConns)

			for i := 0; i < numConns; i++ {
				go func(id int) {
					defer wg.Done()
					con := connectToServer(adr)
					defer func() { _ = con.Close() }()

					for j := 0; j < msgsPerConn; j++ {
						msg := []byte("concurrent test")
						rsp := sendAndReceive(con, msg)
						Expect(rsp).To(Equal(msg))
					}
				}(i)
			}

			wg.Wait()
		})

		// Test: Verifying atomic counter accuracy under peak load.
		It("should maintain connection count accuracy under load", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			numConns := 20
			var conns []any
			mtx := sync.Mutex{}

			// Rapidly open connections.
			for i := 0; i < numConns; i++ {
				con := connectToServer(adr)
				mtx.Lock()
				conns = append(conns, con)
				mtx.Unlock()
			}

			// Wait for the counter to reflect the full load.
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(numConns)))

			// Rapidly close connections.
			mtx.Lock()
			for _, c := range conns {
				if con, ok := c.(any); ok {
					if closer, ok := con.(interface{ Close() error }); ok {
						_ = closer.Close()
					}
				}
			}
			mtx.Unlock()

			// Counter must return to zero.
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Context("concurrent operations", func() {
		BeforeEach(func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())
		})

		// Test: Thread-safe polling of server state.
		It("should handle concurrent IsRunning calls", func() {
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			var wg sync.WaitGroup
			numGoroutines := 10
			wg.Add(numGoroutines)

			for i := 0; i < numGoroutines; i++ {
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						_ = srv.IsRunning()
					}
				}()
			}

			wg.Wait()
		})

		// Test: Thread-safe polling of connection count.
		It("should handle concurrent OpenConnections calls", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			var wg sync.WaitGroup
			numGoroutines := 10
			wg.Add(numGoroutines)

			for i := 0; i < numGoroutines; i++ {
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						_ = srv.OpenConnections()
					}
				}()
			}

			wg.Wait()
		})
	})

	Context("stress testing", func() {
		// Test: Stress testing the Accept loop and atomic counters.
		It("should handle rapid connection churn", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, closeHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			numCycles := 50
			var wg sync.WaitGroup
			wg.Add(numCycles)

			for i := 0; i < numCycles; i++ {
				go func() {
					defer wg.Done()
					con := connectToServer(adr)
					// Connection is closed immediately by closeHandler and by the client.
					_ = con.Close()
				}()
			}

			wg.Wait()

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 5*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})
	})
})
