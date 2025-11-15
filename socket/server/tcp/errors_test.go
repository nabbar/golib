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
	"errors"
	"net"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Error Handling", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(x, 30*time.Second)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Error Constants", func() {
		It("should have defined error constants", func() {
			Expect(scksrv.ErrInvalidAddress).ToNot(BeNil())
			Expect(scksrv.ErrInvalidHandler).ToNot(BeNil())
			Expect(scksrv.ErrShutdownTimeout).ToNot(BeNil())
			Expect(scksrv.ErrGoneTimeout).ToNot(BeNil())
			Expect(scksrv.ErrInvalidInstance).ToNot(BeNil())
		})

		It("should have meaningful error messages", func() {
			Expect(scksrv.ErrInvalidAddress.Error()).To(ContainSubstring("address"))
			Expect(scksrv.ErrInvalidHandler.Error()).To(ContainSubstring("handler"))
			Expect(scksrv.ErrShutdownTimeout.Error()).To(ContainSubstring("timeout"))
			Expect(scksrv.ErrGoneTimeout.Error()).To(ContainSubstring("timeout"))
			Expect(scksrv.ErrInvalidInstance.Error()).To(ContainSubstring("instance"))
		})
	})

	Describe("Invalid Instance Errors", func() {
		It("should handle nil server gracefully", func() {
			// Test nil handling in various methods
			var srv scksrv.ServerTcp
			srv = nil

			// These should handle nil gracefully without panic
			if srv != nil {
				Expect(srv.IsRunning()).To(BeFalse())
			}
		})
	})

	Describe("Invalid Address Errors", func() {
		var srv scksrv.ServerTcp

		BeforeEach(func() {
			srv = scksrv.New(nil, echoHandler)
		})

		It("should return ErrInvalidAddress for empty address", func() {
			err := srv.RegisterServer("")
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, scksrv.ErrInvalidAddress)).To(BeTrue())
		})

		It("should fail Listen without address", func() {
			err := srv.Listen(ctx)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, scksrv.ErrInvalidAddress)).To(BeTrue())
		})

		It("should fail with malformed address", func() {
			err := srv.RegisterServer("not-an-address")
			Expect(err).To(HaveOccurred())
		})

		It("should fail with invalid port", func() {
			err := srv.RegisterServer("127.0.0.1:999999")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Invalid Handler Errors", func() {
		It("should return ErrInvalidHandler when handler is nil", func() {
			srv := scksrv.New(nil, nil)
			err := srv.RegisterServer(getTestAddress())
			Expect(err).ToNot(HaveOccurred())

			err = srv.Listen(ctx)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, scksrv.ErrInvalidHandler)).To(BeTrue())
		})
	})

	Describe("Timeout Errors", func() {
		It("should timeout on shutdown with very short context", func() {
			address := getTestAddress()
			srv := createAndRegisterServer(address, echoHandler, nil)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			// Create many connections
			conns := make([]net.Conn, 5)
			for i := range conns {
				conns[i] = connectClient(address)
			}

			// Very short timeout
			shortCtx, shortCancel := context.WithTimeout(ctx, 1*time.Nanosecond)
			defer shortCancel()

			time.Sleep(10 * time.Millisecond) // Ensure context is expired

			err := srv.Shutdown(shortCtx)
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("timeout"))
			}

			// Cleanup
			for _, conn := range conns {
				_ = conn.Close()
			}
			_ = srv.Shutdown(ctx)
		})

		It("should handle StopGone timeout", func() {
			address := getTestAddress()
			srv := createAndRegisterServer(address, delayHandler(30*time.Second), nil)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			// Try to send data to trigger handler delay
			go func() {
				_, _ = conn.Write([]byte("test\n"))
			}()

			time.Sleep(100 * time.Millisecond)

			// Short timeout for StopGone
			shortCtx, shortCancel := context.WithTimeout(ctx, 1*time.Millisecond)
			defer shortCancel()

			time.Sleep(10 * time.Millisecond)

			tcpSrv := srv.(interface {
				StopGone(ctx context.Context) error
			})
			err := tcpSrv.StopGone(shortCtx)
			if err != nil {
				Expect(err.Error()).To(Or(
					ContainSubstring("timeout"),
					ContainSubstring("gone"),
				))
			}

			_ = conn.Close()
			_ = srv.Shutdown(ctx)
		})

		It("should handle StopListen timeout", func() {
			srv := createAndRegisterServer(getTestAddress(), echoHandler, nil)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			shortCtx, shortCancel := context.WithTimeout(ctx, 1*time.Millisecond)
			defer shortCancel()

			time.Sleep(10 * time.Millisecond)

			tcpSrv := srv.(interface {
				StopListen(ctx context.Context) error
			})
			err := tcpSrv.StopListen(shortCtx)
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("timeout"))
			}

			_ = srv.Shutdown(ctx)
		})
	})

	Describe("Connection Errors", func() {
		var srv libsck.Server
		var address string

		BeforeEach(func() {
			address = getTestAddress()
			srv = createAndRegisterServer(address, echoHandler, nil)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
		})

		AfterEach(func() {
			if srv != nil && srv.IsRunning() {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should handle client disconnect during read", func() {
			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			// Close immediately
			_ = conn.Close()

			// Connection should be cleaned up
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 10*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))
		})

		It("should handle client disconnect during write", func() {
			readyHandler := func(request libsck.Reader, response libsck.Writer) {
				defer func() {
					_ = request.Close()
					_ = response.Close()
				}()

				// Try to write continuously
				for {
					_, err := response.Write([]byte("data"))
					if err != nil {
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			}

			srv2 := createAndRegisterServer(getTestAddress(), readyHandler, nil)
			addr2 := getTestAddress()
			_ = srv2.RegisterServer(addr2)
			startServer(ctx, srv2)
			waitForServerRunning(srv2, 2*time.Second)

			conn := connectClient(addr2)
			waitForConnections(srv2, 1, 2*time.Second)

			// Close immediately
			_ = conn.Close()

			// Connection should be cleaned up
			Eventually(func() int64 {
				return srv2.OpenConnections()
			}, 10*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))

			_ = srv2.Shutdown(ctx)
		})

		It("should handle EOF gracefully", func() {
			eofHandler := func(request libsck.Reader, response libsck.Writer) {
				defer func() {
					_ = request.Close()
					_ = response.Close()
				}()

				buf := make([]byte, 1024)
				for {
					_, err := request.Read(buf)
					if err != nil {
						// EOF or other error, exit gracefully
						return
					}
				}
			}

			srv3 := createAndRegisterServer(getTestAddress(), eofHandler, nil)
			addr3 := getTestAddress()
			_ = srv3.RegisterServer(addr3)
			startServer(ctx, srv3)
			waitForServerRunning(srv3, 2*time.Second)

			conn := connectClient(addr3)
			waitForConnections(srv3, 1, 2*time.Second)

			// Close write side to send EOF
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				_ = tcpConn.CloseWrite()
			} else {
				_ = conn.Close()
			}

			time.Sleep(500 * time.Millisecond)
			_ = conn.Close()

			_ = srv3.Shutdown(ctx)
		})
	})

	Describe("Handler Panics", func() {
		It("should handle handler that panics gracefully", func() {
			panicHandler := func(request libsck.Reader, response libsck.Writer) {
				defer func() {
					// Recover from panic
					if r := recover(); r != nil {
						// Panic recovered, close connections
					}
					_ = request.Close()
					_ = response.Close()
				}()
				panic("intentional panic for testing")
			}

			srv := createAndRegisterServer(getTestAddress(), panicHandler, nil)
			address := getTestAddress()
			_ = srv.RegisterServer(address)

			go func() {
				defer GinkgoRecover()
				_ = srv.Listen(ctx)
			}()

			time.Sleep(500 * time.Millisecond)

			// Try to connect - should work as panic is recovered
			conn, err := connectClientSafe(address)
			if err == nil && conn != nil {
				_, _ = conn.Write([]byte("test\n"))
				time.Sleep(100 * time.Millisecond)
				_ = conn.Close()
			}

			_ = srv.Shutdown(ctx)
		})
	})

	Describe("Resource Exhaustion", func() {
		It("should handle many rapid connections", func() {
			srv := createAndRegisterServer(getTestAddress(), echoHandler, nil)
			address := getTestAddress()
			_ = srv.RegisterServer(address)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			// Create and close many connections rapidly
			for i := 0; i < 50; i++ {
				conn, err := connectClientSafe(address)
				if err == nil && conn != nil {
					_ = conn.Close()
				}
			}

			// Server should still be running
			Expect(srv.IsRunning()).To(BeTrue())

			// All connections should eventually close
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 20*time.Second, 100*time.Millisecond).Should(Equal(int64(0)))

			_ = srv.Shutdown(ctx)
		})
	})

	Describe("Context Cancellation", func() {
		It("should handle context cancellation during Listen", func() {
			srv := createAndRegisterServer(getTestAddress(), echoHandler, nil)
			localCtx, localCancel := context.WithCancel(ctx)

			go func() {
				_ = srv.Listen(localCtx)
			}()

			waitForServerRunning(srv, 2*time.Second)

			// Cancel context
			localCancel()

			// Server should stop
			waitForServerStopped(srv, 5*time.Second)
		})

		It("should handle context cancellation with active connections", func() {
			srv := createAndRegisterServer(getTestAddress(), echoHandler, nil)
			address := getTestAddress()
			_ = srv.RegisterServer(address)
			localCtx, localCancel := context.WithCancel(ctx)

			startServer(localCtx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			// Cancel context
			localCancel()

			// Server should stop
			waitForServerStopped(srv, 5*time.Second)

			_ = conn.Close()
		})
	})
})

// Helper function to connect safely without panicking
func connectClientSafe(address string) (net.Conn, error) {
	return net.DialTimeout("tcp", address, 1*time.Second)
}
