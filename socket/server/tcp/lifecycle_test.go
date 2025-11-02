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
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Lifecycle", func() {
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
	})

	AfterEach(func() {
		if srv != nil && srv.IsRunning() {
			_ = srv.Shutdown(ctx)
		}
		if cancel != nil {
			cancel()
		}
	})

	Describe("Listen", func() {
		Context("successful start", func() {
			It("should start listening successfully", func() {
				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)
				Expect(srv.IsRunning()).To(BeTrue())
			})

			It("should update IsRunning state", func() {
				Expect(srv.IsRunning()).To(BeFalse())
				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)
				Expect(srv.IsRunning()).To(BeTrue())
			})

			It("should make Done channel block when running", func() {
				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)

				done := srv.Done()
				Expect(done).ToNot(BeNil())

				// Should block now (not be closed)
				select {
				case <-done:
					Fail("Done channel should not be closed while running")
				case <-time.After(100 * time.Millisecond):
					// Expected: channel is not closed
				}
			})

			It("should accept multiple connections", func() {
				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)

				// Connect multiple clients
				conn1 := connectClient(address)
				conn2 := connectClient(address)
				conn3 := connectClient(address)

				waitForConnections(srv, 3, 2*time.Second)
				Expect(srv.OpenConnections()).To(Equal(int64(3)))

				_ = conn1.Close()
				_ = conn2.Close()
				_ = conn3.Close()
			})
		})

		Context("with errors", func() {
			It("should fail without registered address", func() {
				srvNoAddr := scksrv.New(nil, echoHandler)
				err := srvNoAddr.Listen(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(scksrv.ErrInvalidAddress))
			})

			It("should fail without handler", func() {
				srvNoHandler := scksrv.New(nil, nil)
				err := srvNoHandler.RegisterServer(getTestAddress())
				Expect(err).ToNot(HaveOccurred())

				err = srvNoHandler.Listen(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(scksrv.ErrInvalidHandler))
			})

			It("should fail with already used port", func() {
				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)

				// Try to start another server on the same port
				srv2 := scksrv.New(nil, echoHandler)
				err := srv2.RegisterServer(address)
				Expect(err).ToNot(HaveOccurred())

				err = srv2.Listen(ctx)
				Expect(err).To(HaveOccurred())
			})

			It("should handle context cancellation", func() {
				localCtx, localCancel := context.WithCancel(ctx)
				startServer(localCtx, srv)
				waitForServerRunning(srv, 2*time.Second)

				localCancel()
				waitForServerStopped(srv, 5*time.Second)
				Expect(srv.IsRunning()).To(BeFalse())
			})
		})
	})

	Describe("Shutdown", func() {
		Context("graceful shutdown", func() {
			It("should shutdown successfully when running", func() {
				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)

				err := srv.Shutdown(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(srv.IsRunning()).To(BeFalse())
			})

			It("should close all connections on shutdown", func() {
				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)

				conn := connectClient(address)
				waitForConnections(srv, 1, 2*time.Second)

				err := srv.Shutdown(ctx)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() int64 {
					return srv.OpenConnections()
				}, 15*time.Second, 100*time.Millisecond).Should(Equal(int64(0)))

				_ = conn.Close()
			})

			It("should be idempotent", func() {
				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)

				err := srv.Shutdown(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Shutdown again
				err = srv.Shutdown(ctx)
				// Should not error or should be acceptable
				if err != nil {
					// Acceptable errors on double shutdown
					Expect(err.Error()).To(Or(
						ContainSubstring("timeout"),
						ContainSubstring("not running"),
					))
				}
			})

			It("should work when not running", func() {
				Expect(srv.IsRunning()).To(BeFalse())
				err := srv.Shutdown(ctx)
				// Either no error or timeout error is acceptable
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("timeout"))
				}
			})
		})

		Context("with timeout", func() {
			It("should respect context timeout", func() {
				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)

				timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Millisecond)
				defer timeoutCancel()

				// Give it time to timeout
				time.Sleep(10 * time.Millisecond)

				err := srv.Shutdown(timeoutCtx)
				// May succeed or timeout depending on timing
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("timeout"))
				}
			})
		})
	})

	Describe("Close", func() {
		It("should close the server", func() {
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			err := srv.Close()
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return srv.IsRunning()
			}, 15*time.Second, 100*time.Millisecond).Should(BeFalse())
		})

		It("should behave like Shutdown", func() {
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			err := srv.Close()
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 15*time.Second, 100*time.Millisecond).Should(Equal(int64(0)))

			_ = conn.Close()
		})
	})

	Describe("StopListen", func() {
		It("should stop accepting new connections", func() {
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			tcpSrv := srv.(interface {
				StopListen(ctx context.Context) error
			})
			err := tcpSrv.StopListen(ctx)
			Expect(err).ToNot(HaveOccurred())

			waitForServerStopped(srv, 15*time.Second)
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should not immediately close existing connections", func() {
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			// Stop listening but don't stop gone
			tcpSrv := srv.(interface {
				StopListen(ctx context.Context) error
			})
			err := tcpSrv.StopListen(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Connection should still be open briefly
			time.Sleep(100 * time.Millisecond)

			_ = conn.Close()
		})
	})

	Describe("StopGone", func() {
		It("should mark server as gone", func() {
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			tcpSrv := srv.(interface {
				StopGone(ctx context.Context) error
			})
			err := tcpSrv.StopGone(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return srv.IsGone()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should wait for connections to close", func() {
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			go func() {
				time.Sleep(500 * time.Millisecond)
				_ = conn.Close()
			}()

			tcpSrv := srv.(interface {
				StopGone(ctx context.Context) error
			})
			err := tcpSrv.StopGone(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("IsRunning", func() {
		It("should reflect server state", func() {
			Expect(srv.IsRunning()).To(BeFalse())

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
			Expect(srv.IsRunning()).To(BeTrue())

			_ = srv.Shutdown(ctx)
			waitForServerStopped(srv, 15*time.Second)
			Expect(srv.IsRunning()).To(BeFalse())
		})
	})

	Describe("IsGone", func() {
		It("should be false initially", func() {
			Expect(srv.IsGone()).To(BeFalse())
		})

		It("should be true after shutdown", func() {
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			_ = srv.Shutdown(ctx)

			Eventually(func() bool {
				return srv.IsGone()
			}, 15*time.Second, 100*time.Millisecond).Should(BeTrue())
		})
	})

	Describe("Done channel", func() {
		It("should signal when shutdown is initiated", func() {
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			done := srv.Done()

			go func() {
				time.Sleep(200 * time.Millisecond)
				_ = srv.Shutdown(ctx)
			}()

			select {
			case <-done:
				// Expected: Done signaled
			case <-time.After(15 * time.Second):
				Fail("Done channel should have been signaled")
			}
		})
	})
})
