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

package tcp_test

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Connections", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		srv     libsck.Server
		address string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(x, 60*time.Second)
		address = getTestAddress()
		srv = createAndRegisterServer(address, echoHandler, nil)
		startServer(ctx, srv)
		waitForServerRunning(srv, 5*time.Second)
	})

	AfterEach(func() {
		if srv != nil && srv.IsRunning() {
			_ = srv.Shutdown(ctx)
		}
		if cancel != nil {
			cancel()
		}
	})

	Describe("Single Connection", func() {
		It("should accept a single connection", func() {
			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)
			Expect(srv.OpenConnections()).To(Equal(int64(1)))
		})

		It("should echo data back to client", func() {
			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)

			msg := []byte("Hello, World!\n")
			n := sendMessage(conn, msg)
			Expect(n).To(Equal(len(msg)))

			response := receiveMessage(conn, 1024)
			Expect(response).To(Equal(msg))
		})

		It("should handle multiple messages on same connection", func() {
			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)

			for i := 0; i < 5; i++ {
				msg := []byte("Message number " + string(rune('0'+i)) + "\n")
				sendMessage(conn, msg)
				response := receiveMessage(conn, 1024)
				Expect(response).To(Equal(msg))
			}
		})

		It("should decrement connection count on disconnect", func() {
			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			_ = conn.Close()

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 10*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Describe("Multiple Connections", func() {
		It("should handle multiple simultaneous connections", func() {
			conn1 := connectClient(address)
			defer func() {
				_ = conn1.Close()
			}()

			conn2 := connectClient(address)
			defer func() {
				_ = conn2.Close()
			}()

			conn3 := connectClient(address)
			defer func() {
				_ = conn3.Close()
			}()

			waitForConnections(srv, 3, 2*time.Second)
			Expect(srv.OpenConnections()).To(Equal(int64(3)))
		})

		It("should handle each connection independently", func() {
			conn1 := connectClient(address)
			defer func() {
				_ = conn1.Close()
			}()

			conn2 := connectClient(address)
			defer func() {
				_ = conn2.Close()
			}()

			waitForConnections(srv, 2, 2*time.Second)

			msg1 := []byte("Connection 1\n")
			msg2 := []byte("Connection 2\n")

			sendMessage(conn1, msg1)
			sendMessage(conn2, msg2)

			resp1 := receiveMessage(conn1, 1024)
			resp2 := receiveMessage(conn2, 1024)

			Expect(resp1).To(Equal(msg1))
			Expect(resp2).To(Equal(msg2))
		})

		It("should track connection count correctly", func() {
			// Connect 3 clients
			conn1 := connectClient(address)
			conn2 := connectClient(address)
			conn3 := connectClient(address)

			waitForConnections(srv, 3, 2*time.Second)

			// Close one
			_ = conn1.Close()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 10*time.Second, 50*time.Millisecond).Should(Equal(int64(2)))

			// Close another
			_ = conn2.Close()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 10*time.Second, 50*time.Millisecond).Should(Equal(int64(1)))

			// Close last
			_ = conn3.Close()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 10*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))
		})

		It("should handle rapid connections and disconnections", func() {
			for i := 0; i < 10; i++ {
				conn := connectClient(address)
				_ = conn.Close()
			}

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 15*time.Second, 100*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Describe("Connection Lifecycle", func() {
		It("should handle connection read closure", func() {
			conn := connectClient(address)

			waitForConnections(srv, 1, 2*time.Second)

			// Send a message
			msg := []byte("Test\n")
			sendMessage(conn, msg)
			response := receiveMessage(conn, 1024)
			Expect(response).To(Equal(msg))

			// Close read side if possible
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				_ = tcpConn.CloseRead()
			}

			// Close the connection fully
			_ = conn.Close()

			// Connection should eventually close
			time.Sleep(1 * time.Second)
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 15*time.Second, 100*time.Millisecond).Should(Equal(int64(0)))
		})

		It("should handle connection write closure", func() {
			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)

			// Send a message
			msg := []byte("Test\n")
			sendMessage(conn, msg)
			response := receiveMessage(conn, 1024)
			Expect(response).To(Equal(msg))

			// Close write side if possible
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				_ = tcpConn.CloseWrite()
			}

			// Connection should eventually close
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 10*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Describe("UpdateConn Callback", func() {
		It("should call UpdateConn for each connection", func() {
			updateCount := &atomic.Int32{}
			updateConn := func(conn net.Conn) {
				updateCount.Add(1)
				Expect(conn).ToNot(BeNil())
			}

			srvWithUpdate := createAndRegisterServer(getTestAddress(), echoHandler, updateConn)
			defer func() {
				_ = srvWithUpdate.Shutdown(ctx)
			}()

			addr := getTestAddress()
			_ = srvWithUpdate.RegisterServer(addr)
			startServer(ctx, srvWithUpdate)
			waitForServerRunning(srvWithUpdate, 2*time.Second)

			// Connect multiple clients
			conn1 := connectClient(addr)
			defer func() {
				_ = conn1.Close()
			}()

			conn2 := connectClient(addr)
			defer func() {
				_ = conn2.Close()
			}()

			conn3 := connectClient(addr)
			defer func() {
				_ = conn3.Close()
			}()

			Eventually(func() int32 {
				return updateCount.Load()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int32(3)))
		})

		It("should allow connection modification", func() {
			updateConn := func(conn net.Conn) {
				// Set options on the connection
				if tcpConn, ok := conn.(*net.TCPConn); ok {
					_ = tcpConn.SetKeepAlive(true)
					_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
					_ = tcpConn.SetNoDelay(true)
				}
			}

			srvWithUpdate := createAndRegisterServer(getTestAddress(), echoHandler, updateConn)
			defer func() {
				_ = srvWithUpdate.Shutdown(ctx)
			}()

			addr := getTestAddress()
			_ = srvWithUpdate.RegisterServer(addr)
			startServer(ctx, srvWithUpdate)
			waitForServerRunning(srvWithUpdate, 2*time.Second)

			conn := connectClient(addr)
			defer func() {
				_ = conn.Close()
			}()

			// Should work normally
			msg := []byte("Test\n")
			sendMessage(conn, msg)
			response := receiveMessage(conn, 1024)
			Expect(response).To(Equal(msg))
		})
	})

	Describe("Large Data Transfer", func() {
		It("should handle large messages", func() {
			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)

			// Create a large message (1MB)
			largeMsg := make([]byte, 1024*1024)
			for i := range largeMsg {
				largeMsg[i] = byte(i % 256)
			}

			// Send in chunks
			go func() {
				_, _ = conn.Write(largeMsg)
			}()

			// Receive in chunks
			received := make([]byte, 0, len(largeMsg))
			buf := make([]byte, 8192)
			deadline := time.Now().Add(10 * time.Second)
			_ = conn.SetReadDeadline(deadline)

			for len(received) < len(largeMsg) {
				n, err := conn.Read(buf)
				if err != nil {
					if err == io.EOF {
						break
					}
					Expect(err).ToNot(HaveOccurred())
				}
				received = append(received, buf[:n]...)
			}

			Expect(len(received)).To(Equal(len(largeMsg)))
			Expect(received).To(Equal(largeMsg))
		})

		It("should handle concurrent data on multiple connections", func() {
			const numConns = 5
			const msgSize = 10 * 1024 // 10KB

			conns := make([]net.Conn, numConns)
			for i := 0; i < numConns; i++ {
				conns[i] = connectClient(address)
				defer func() {
					_ = conns[i].Close()
				}()
			}

			waitForConnections(srv, int64(numConns), 2*time.Second)

			// Send and receive on all connections concurrently
			done := make(chan bool, numConns)

			for i := 0; i < numConns; i++ {
				go func(connIdx int) {
					defer GinkgoRecover()
					conn := conns[connIdx]

					msg := make([]byte, msgSize)
					for j := range msg {
						msg[j] = byte((connIdx + j) % 256)
					}

					// Send
					_, err := conn.Write(msg)
					Expect(err).ToNot(HaveOccurred())

					// Receive
					received := make([]byte, msgSize)
					_, err = io.ReadFull(conn, received)
					Expect(err).ToNot(HaveOccurred())
					Expect(received).To(Equal(msg))

					done <- true
				}(i)
			}

			// Wait for all to complete
			for i := 0; i < numConns; i++ {
				select {
				case <-done:
					// Success
				case <-time.After(10 * time.Second):
					Fail("Timeout waiting for concurrent connections")
				}
			}
		})
	})

	Describe("Connection Timeouts", func() {
		It("should handle slow clients", func() {
			slowHandler := func(request libsck.Reader, response libsck.Writer) {
				defer func() {
					_ = request.Close()
					_ = response.Close()
				}()
				// Read slowly
				buf := make([]byte, 1)
				for {
					_, err := request.Read(buf)
					if err != nil {
						return
					}
					time.Sleep(10 * time.Millisecond)
					_, _ = response.Write(buf)
				}
			}

			slowSrv := createAndRegisterServer(getTestAddress(), slowHandler, nil)
			defer func() {
				_ = slowSrv.Shutdown(ctx)
			}()

			addr := getTestAddress()
			_ = slowSrv.RegisterServer(addr)
			startServer(ctx, slowSrv)
			waitForServerRunning(slowSrv, 2*time.Second)

			conn := connectClient(addr)
			defer func() {
				_ = conn.Close()
			}()

			// Send message
			msg := []byte("slow")
			sendMessage(conn, msg)

			// Should eventually receive response
			buf := make([]byte, len(msg))
			_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			_, err := io.ReadFull(conn, buf)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
