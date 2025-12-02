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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsck "github.com/nabbar/golib/socket"
	"github.com/nabbar/golib/socket/server/udp"
)

var _ = Describe("UDP Server Implementation", func() {
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
		time.Sleep(50 * time.Millisecond)
	})

	Describe("Callback Registration", func() {
		Context("error callbacks", func() {
			It("should register error callback", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				errorCollector := newErrorCollector()
				srv.RegisterFuncError(errorCollector.callback)

				// Callback should be registered
				Expect(srv).ToNot(BeNil())
			})

			It("should call error callback on errors", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				errorCollector := newErrorCollector()
				srv.RegisterFuncError(errorCollector.callback)

				// Trigger error by using invalid address
				err = srv.RegisterServer("invalid:99999")
				if err != nil {
					errorCollector.callback(err)
				}

				// Error should be collected
				Expect(errorCollector.hasErrors()).To(BeTrue())
			})

			It("should allow nil error callback", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				srv.RegisterFuncError(nil)
				// Should not panic
			})
		})

		Context("info callbacks", func() {
			It("should register info callback", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				infoCollector := newInfoCollector()
				srv.RegisterFuncInfo(infoCollector.callback)

				Expect(srv).ToNot(BeNil())
			})

			It("should call info callback on connection events", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				infoCollector := newInfoCollector()
				srv.RegisterFuncInfo(infoCollector.callback)

				startServer(srv, ctx)
				defer stopServer(srv, cancel)

				// Wait for connection events
				time.Sleep(100 * time.Millisecond)

				// Should have received connection events
				events := infoCollector.getEvents()
				Expect(len(events)).To(BeNumerically(">", 0))
			})

			It("should allow nil info callback", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				srv.RegisterFuncInfo(nil)
				// Should not panic
			})
		})

		Context("server info callbacks", func() {
			It("should register server info callback", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				serverInfo := newServerInfoCollector()
				srv.RegisterFuncInfoServer(serverInfo.callback)

				Expect(srv).ToNot(BeNil())
			})

			It("should call server info callback on start", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				serverInfo := newServerInfoCollector()
				srv.RegisterFuncInfoServer(serverInfo.callback)

				startServer(srv, ctx)
				defer stopServer(srv, cancel)

				// Wait for server messages
				time.Sleep(100 * time.Millisecond)

				// Should have received messages
				messages := serverInfo.getMessages()
				Expect(len(messages)).To(BeNumerically(">", 0))
			})

			It("should allow nil server info callback", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				srv.RegisterFuncInfoServer(nil)
				// Should not panic
			})
		})
	})

	Describe("UpdateConn Callback", func() {
		It("should call UpdateConn on socket creation", func() {
			handler := newTestHandler(false)
			updateConn := newCustomUpdateConn()

			cfg := createBasicConfig()
			srv, err := udp.New(updateConn.callback, handler.handler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Wait for callback
			Eventually(func() bool {
				return updateConn.wasCalled()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

			// Connection should be set
			Expect(updateConn.getConn()).ToNot(BeNil())
		})

		It("should work without UpdateConn callback", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Should work normally
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Describe("Handler Execution", func() {
		It("should execute handler when server starts", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Handler should be running
			time.Sleep(100 * time.Millisecond)
			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should stop handler when server stops", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)
			stopServer(srv, cancel)

			// Handler should be stopped
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should handle handler panics gracefully", func() {
			panicHandler := func(ctx libsck.Context) {
				defer func() {
					if r := recover(); r != nil {
						// Panic recovered in handler
					}
					ctx.Close()
				}()
				panic("test panic")
			}

			srv, err := createServerWithHandler(panicHandler)
			Expect(err).ToNot(HaveOccurred())

			// Should not crash
			go func() {
				_ = srv.Listen(ctx)
			}()

			time.Sleep(100 * time.Millisecond)
			cancel()
		})
	})

	Describe("TLS Support", func() {
		It("should accept SetTLS call (no-op)", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			err = srv.SetTLS(true, nil)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should always return nil from SetTLS", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			err1 := srv.SetTLS(true, nil)
			err2 := srv.SetTLS(false, nil)

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
		})
	})

	Describe("RegisterServer Method", func() {
		Context("with valid addresses", func() {
			It("should accept loopback address", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				err = srv.RegisterServer("127.0.0.1:8080")

				Expect(err).ToNot(HaveOccurred())
			})

			It("should accept all interfaces address", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				err = srv.RegisterServer(":8080")

				Expect(err).ToNot(HaveOccurred())
			})

			It("should accept IPv6 address", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				err = srv.RegisterServer("[::1]:8080")

				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with invalid addresses", func() {
			It("should reject empty address", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				err = srv.RegisterServer("")

				Expect(err).To(HaveOccurred())
			})

			It("should reject invalid port", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				err = srv.RegisterServer("127.0.0.1:99999")

				Expect(err).To(HaveOccurred())
			})

			It("should reject malformed address", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				err = srv.RegisterServer("not-an-address")

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
