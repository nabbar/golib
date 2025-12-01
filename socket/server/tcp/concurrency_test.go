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

// concurrency_test.go validates the thread-safety and concurrent operation
// of the TCP server. Tests include race condition detection, parallel connection
// handling, and concurrent state management using the race detector.
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

	BeforeEach(func() {
		adr = getTestAddr()
		c, cnl = context.WithCancel(globalCtx)
	})

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
			Expect(cnt.Load()).To(BeNumerically(">=", int32(numConns)))
		})

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

			for i := 0; i < numConns; i++ {
				con := connectToServer(adr)
				mtx.Lock()
				conns = append(conns, con)
				mtx.Unlock()
			}

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(numConns)))

			mtx.Lock()
			for _, c := range conns {
				if con, ok := c.(any); ok {
					if closer, ok := con.(interface{ Close() error }); ok {
						_ = closer.Close()
					}
				}
			}
			mtx.Unlock()

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
