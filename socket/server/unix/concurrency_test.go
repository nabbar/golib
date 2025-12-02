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

// concurrency_test.go validates concurrent server operations and thread safety.
// Tests multiple simultaneous connections, parallel operations, and race conditions.
package unix_test

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	scksru "github.com/nabbar/golib/socket/server/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unix Server Concurrency", func() {
	var (
		srv        scksru.ServerUnix
		socketPath string
		c          context.Context
		cnl        context.CancelFunc
	)

	BeforeEach(func() {
		socketPath = getTestSocketPath()
		c, cnl = context.WithCancel(globalCtx)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		cleanupSocketFile(socketPath)
		time.Sleep(100 * time.Millisecond)
	})

	Context("multiple concurrent connections", func() {
		It("should handle 10 concurrent connections", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			const numConns = 10
			wg := sync.WaitGroup{}
			wg.Add(numConns)

			for i := 0; i < numConns; i++ {
				go func(id int) {
					defer wg.Done()

					con := connectToServer(socketPath)
					defer func() { _ = con.Close() }()

					data := []byte("Message from connection " + string(rune('0'+id)))
					rsp := sendAndReceive(con, data)
					Expect(rsp).To(Equal(data))
				}(i)
			}

			wg.Wait()

			// All connections should eventually close
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})

		It("should handle 50 concurrent connections", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			const numConns = 50
			wg := sync.WaitGroup{}
			wg.Add(numConns)
			successCount := &atomic.Int32{}

			for i := 0; i < numConns; i++ {
				go func(id int) {
					defer wg.Done()

					con := connectToServer(socketPath)
					defer func() { _ = con.Close() }()

					data := []byte("Test data")
					rsp := sendAndReceive(con, data)
					if string(rsp) == string(data) {
						successCount.Add(1)
					}
				}(i)
			}

			wg.Wait()
			Expect(successCount.Load()).To(Equal(int32(numConns)))
		})
	})

	Context("connection count tracking under concurrency", func() {
		It("should maintain accurate connection count", func() {
			cnt := &atomic.Int32{}
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, counterHandler(cnt), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			const numConns = 20
			wg := sync.WaitGroup{}
			wg.Add(numConns)

			for i := 0; i < numConns; i++ {
				go func() {
					defer wg.Done()

					con := connectToServer(socketPath)
					defer func() { _ = con.Close() }()

					data := []byte("Test")
					_ = sendAndReceive(con, data)
				}()
			}

			wg.Wait()
			Expect(cnt.Load()).To(BeNumerically(">=", int32(numConns)))
		})
	})

	Context("rapid connect/disconnect", func() {
		It("should handle rapid connection cycles", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, closeHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			const numCycles = 30
			wg := sync.WaitGroup{}
			wg.Add(numCycles)

			for i := 0; i < numCycles; i++ {
				go func() {
					defer wg.Done()

					con := connectToServer(socketPath)
					_ = con.Close()
				}()
			}

			wg.Wait()

			// All connections should be closed
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Context("concurrent server operations", func() {
		It("should handle concurrent IsRunning calls", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			wg := sync.WaitGroup{}
			wg.Add(10)

			for i := 0; i < 10; i++ {
				go func() {
					defer wg.Done()
					_ = srv.IsRunning()
				}()
			}

			wg.Wait()
		})

		It("should handle concurrent OpenConnections calls", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			wg := sync.WaitGroup{}
			wg.Add(20)

			for i := 0; i < 20; i++ {
				go func() {
					defer wg.Done()
					_ = srv.OpenConnections()
				}()
			}

			wg.Wait()
		})
	})

	Context("concurrent callback registration", func() {
		It("should handle concurrent RegisterFuncError", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			wg := sync.WaitGroup{}
			wg.Add(10)

			for i := 0; i < 10; i++ {
				go func() {
					defer wg.Done()
					srv.RegisterFuncError(func(errs ...error) {})
				}()
			}

			wg.Wait()
		})
	})

	Context("shutdown under load", func() {
		It("should shutdown cleanly with active connections", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, slowHandler(200*time.Millisecond), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			// Create several connections
			conns := make([]interface{}, 5)
			for i := 0; i < 5; i++ {
				conns[i] = connectToServer(socketPath)
			}

			// Shutdown while connections are active
			err = srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			// Cleanup
			for _, con := range conns {
				if con != nil {
					_ = con.(interface{ Close() error }).Close()
				}
			}
		})
	})
})
