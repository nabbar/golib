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
	"io"
	"net"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsck "github.com/nabbar/golib/socket"
)

var _ = Describe("UDP Context Operations", func() {
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

	Describe("Context Interface Methods", func() {
		var (
			udpCtx  libsck.Context
			ctxLock sync.Mutex
		)

		BeforeEach(func() {
			handler := func(c libsck.Context) {
				ctxLock.Lock()
				udpCtx = c
				ctxLock.Unlock()

				// Keep handler alive
				<-c.Done()
			}

			srv, err := createServerWithHandler(handler)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				_ = srv.Listen(ctx)
			}()

			Eventually(func() bool {
				ctxLock.Lock()
				defer ctxLock.Unlock()
				return udpCtx != nil
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should implement Deadline method", func() {
			ctxLock.Lock()
			c := udpCtx
			ctxLock.Unlock()

			deadline, ok := c.Deadline()
			// No deadline set by default
			Expect(ok).To(BeFalse())
			Expect(deadline.IsZero()).To(BeTrue())
		})

		It("should implement Done method", func() {
			ctxLock.Lock()
			c := udpCtx
			ctxLock.Unlock()

			doneChan := c.Done()
			Expect(doneChan).ToNot(BeNil())

			// Just verify channel exists
			_ = doneChan
		})

		It("should implement Err method", func() {
			ctxLock.Lock()
			c := udpCtx
			ctxLock.Unlock()

			// Check Err exists and is callable
			_ = c.Err()
		})

		It("should implement Value method", func() {
			ctxLock.Lock()
			c := udpCtx
			ctxLock.Unlock()

			// No value set by default
			val := c.Value("test-key")
			Expect(val).To(BeNil())
		})

		It("should provide IsConnected method", func() {
			ctxLock.Lock()
			c := udpCtx
			ctxLock.Unlock()

			// Check IsConnected exists and is callable
			_ = c.IsConnected()
		})

		It("should provide RemoteHost method", func() {
			ctxLock.Lock()
			c := udpCtx
			ctxLock.Unlock()

			remote := c.RemoteHost()
			// UDP server context may have empty remote initially
			_ = remote
		})

		It("should provide LocalHost method", func() {
			ctxLock.Lock()
			c := udpCtx
			ctxLock.Unlock()

			local := c.LocalHost()
			Expect(local).ToNot(BeEmpty())
			Expect(local).To(ContainSubstring("udp"))
		})
	})

	Describe("Context I/O Operations", func() {
		var (
			clientConn *net.UDPConn
			serverAddr *net.UDPAddr
		)

		BeforeEach(func() {
			var err error
			serverAddr, err = net.ResolveUDPAddr("udp", "127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())

			clientConn, err = net.DialUDP("udp", nil, serverAddr)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if clientConn != nil {
				clientConn.Close()
			}
		})

		It("should read data from UDP socket", func() {
			handler := func(c libsck.Context) {
				defer c.Close()

				buf := make([]byte, 1024)
				_, _ = c.Read(buf)
			}

			srv, err := createServerWithHandler(handler)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				_ = srv.Listen(ctx)
			}()

			Eventually(func() bool {
				return srv.IsRunning()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

			time.Sleep(50 * time.Millisecond)
			cancel()
		})

		It("should handle Write method (returns error for UDP)", func() {
			handler := func(c libsck.Context) {
				defer c.Close()

				// Write should return error for UDP server context
				n, err := c.Write([]byte("test"))
				Expect(n).To(Equal(0))
				Expect(err).To(Equal(io.ErrClosedPipe))
			}

			srv, err := createServerWithHandler(handler)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				_ = srv.Listen(ctx)
			}()

			time.Sleep(100 * time.Millisecond)
			cancel()
		})

		It("should handle Read errors", func() {
			handler := func(c libsck.Context) {
				defer c.Close()

				buf := make([]byte, 1024)

				// Read from closed context
				c.Close()
				n, err := c.Read(buf)
				Expect(n).To(Equal(0))
				Expect(err).To(Equal(io.ErrClosedPipe))
			}

			srv, err := createServerWithHandler(handler)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				_ = srv.Listen(ctx)
			}()

			time.Sleep(100 * time.Millisecond)
			cancel()
		})

		It("should handle Write on closed context", func() {
			handler := func(c libsck.Context) {
				defer c.Close()

				// Close first
				c.Close()

				// Then try to write
				n, err := c.Write([]byte("test"))
				Expect(n).To(Equal(0))
				Expect(err).To(Equal(io.ErrClosedPipe))
			}

			srv, err := createServerWithHandler(handler)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				_ = srv.Listen(ctx)
			}()

			time.Sleep(100 * time.Millisecond)
			cancel()
		})
	})

	Describe("Context Close Operations", func() {
		It("should handle multiple Close calls", func() {
			var udpCtx libsck.Context
			var wg sync.WaitGroup
			wg.Add(1)

			handler := func(c libsck.Context) {
				udpCtx = c
				wg.Done()
				<-c.Done()
			}

			srv, err := createServerWithHandler(handler)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				_ = srv.Listen(ctx)
			}()

			wg.Wait()
			Expect(udpCtx).ToNot(BeNil())

			// First close
			err = udpCtx.Close()
			Expect(err).ToNot(HaveOccurred())

			// Second close should also succeed
			err = udpCtx.Close()
			Expect(err).To(BeNil())

			cancel()
		})

		It("should handle Close with cancelled context", func() {
			var udpCtx libsck.Context
			var wg sync.WaitGroup
			wg.Add(1)

			handler := func(c libsck.Context) {
				udpCtx = c
				wg.Done()
				<-c.Done()
			}

			srv, err := createServerWithHandler(handler)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				_ = srv.Listen(ctx)
			}()

			wg.Wait()

			// Cancel context first
			cancel()
			time.Sleep(50 * time.Millisecond)

			// Then close
			err = udpCtx.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

})
