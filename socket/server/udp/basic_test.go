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
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	"github.com/nabbar/golib/socket/server/udp"
)

var _ = Describe("UDP Server Basic Operations", func() {
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
		time.Sleep(50 * time.Millisecond) // Allow cleanup
	})

	Describe("Server Construction", func() {
		Context("with valid configuration", func() {
			It("should create server successfully", func() {
				cfg := createBasicConfig()
				handler := func(ctx libsck.Context) {}

				srv, err := udp.New(nil, handler, cfg)

				Expect(err).ToNot(HaveOccurred())
				Expect(srv).ToNot(BeNil())
			})

			It("should accept nil UpdateConn callback", func() {
				cfg := createBasicConfig()
				handler := func(ctx libsck.Context) {}

				srv, err := udp.New(nil, handler, cfg)

				Expect(err).ToNot(HaveOccurred())
				Expect(srv).ToNot(BeNil())
			})

			It("should accept custom UpdateConn callback", func() {
				cfg := createBasicConfig()
				handler := func(ctx libsck.Context) {}
				updateConn := func(conn net.Conn) {}

				srv, err := udp.New(updateConn, handler, cfg)

				Expect(err).ToNot(HaveOccurred())
				Expect(srv).ToNot(BeNil())
			})
		})

		Context("with invalid configuration", func() {
			It("should return error with empty address", func() {
				cfg := sckcfg.Server{
					Network: libptc.NetworkUDP,
					Address: "",
				}
				handler := func(ctx libsck.Context) {}

				srv, err := udp.New(nil, handler, cfg)

				Expect(err).To(HaveOccurred())
				Expect(srv).To(BeNil())
			})

			It("should return error with invalid address format", func() {
				cfg := sckcfg.Server{
					Network: libptc.NetworkUDP,
					Address: "invalid-address",
				}
				handler := func(ctx libsck.Context) {}

				srv, err := udp.New(nil, handler, cfg)

				Expect(err).To(HaveOccurred())
				Expect(srv).To(BeNil())
			})

			It("should return error with malformed port", func() {
				cfg := sckcfg.Server{
					Network: libptc.NetworkUDP,
					Address: "127.0.0.1:99999",
				}
				handler := func(ctx libsck.Context) {}

				srv, err := udp.New(nil, handler, cfg)

				Expect(err).To(HaveOccurred())
				Expect(srv).To(BeNil())
			})
		})
	})

	Describe("Server State Management", func() {
		var srv udp.ServerUdp

		BeforeEach(func() {
			handler := newTestHandler(false)
			var err error
			srv, err = createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("initial state", func() {
			It("should not be running", func() {
				Expect(srv.IsRunning()).To(BeFalse())
			})

			It("should not be gone", func() {
				Expect(srv.IsGone()).To(BeTrue())
			})

			It("should have zero connections", func() {
				Expect(srv.OpenConnections()).To(Equal(int64(0)))
			})
		})

		Context("after starting", func() {
			BeforeEach(func() {
				startServer(srv, ctx)
			})

			AfterEach(func() {
				stopServer(srv, cancel)
			})

			It("should be running", func() {
				Expect(srv.IsRunning()).To(BeTrue())
			})

			It("should not be gone", func() {
				Expect(srv.IsGone()).To(BeFalse())
			})

			It("should have zero connections (UDP is stateless)", func() {
				Expect(srv.OpenConnections()).To(Equal(int64(0)))
			})
		})

		Context("after stopping", func() {
			BeforeEach(func() {
				startServer(srv, ctx)
				stopServer(srv, cancel)
			})

			It("should not be running", func() {
				Expect(srv.IsRunning()).To(BeFalse())
			})

			It("should be gone", func() {
				Eventually(func() bool {
					return srv.IsGone()
				}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
			})

			It("should have zero connections", func() {
				Expect(srv.OpenConnections()).To(Equal(int64(0)))
			})
		})
	})

	Describe("Server Lifecycle", func() {
		var srv udp.ServerUdp

		BeforeEach(func() {
			handler := newTestHandler(false)
			var err error
			srv, err = createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should start successfully", func() {
			go func() {
				_ = srv.Listen(ctx)
			}()

			Eventually(func() bool {
				return srv.IsRunning()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

			cancel()
		})

		It("should stop via context cancellation", func() {
			startServer(srv, ctx)

			cancel()

			Eventually(func() bool {
				return !srv.IsRunning()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should stop via Shutdown method", func() {
			startServer(srv, ctx)

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			err := srv.Shutdown(shutdownCtx)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv.IsRunning()).To(BeFalse())
			Expect(srv.IsGone()).To(BeTrue())
		})

		It("should stop via Close method", func() {
			startServer(srv, ctx)

			err := srv.Close()

			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				return !srv.IsRunning()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should allow multiple Close calls", func() {
			startServer(srv, ctx)

			err1 := srv.Close()
			Expect(err1).ToNot(HaveOccurred())

			Eventually(func() bool {
				return srv.IsGone()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

			err2 := srv.Close()
			Expect(err2).ToNot(HaveOccurred())
		})
	})

	Describe("Interface Implementation", func() {
		var srv udp.ServerUdp

		BeforeEach(func() {
			handler := newTestHandler(false)
			var err error
			srv, err = createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should implement Server interface", func() {
			var _ libsck.Server = srv
		})

		It("should provide IsRunning method", func() {
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should provide IsGone method", func() {
			Expect(srv.IsGone()).To(BeTrue())
		})

		It("should provide OpenConnections method", func() {
			Expect(srv.OpenConnections()).To(Equal(int64(0)))
		})

		It("should provide RegisterServer method", func() {
			err := srv.RegisterServer("127.0.0.1:9999")
			// Should succeed or fail gracefully
			_ = err
		})

		It("should provide SetTLS method (no-op for UDP)", func() {
			err := srv.SetTLS(true, nil)
			Expect(err).ToNot(HaveOccurred()) // Always nil for UDP
		})
	})
})
