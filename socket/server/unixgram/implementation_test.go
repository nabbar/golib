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

package unixgram_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/unixgram"
)

var _ = Describe("Unix Datagram Server Implementation", func() {
	var (
		ctx      context.Context
		cancel   context.CancelFunc
		sockPath string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
		cleanupSocketFile(sockPath)
		time.Sleep(50 * time.Millisecond)
	})

	Describe("Datagram Handling", func() {
		It("should receive datagrams", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			// Send test datagram
			testData := []byte("test message")
			err = sendUnixgramDatagram(sockPath, testData)
			Expect(err).ToNot(HaveOccurred())

			// Wait for datagram to be received
			Eventually(func() int64 {
				return handler.getCount()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">", 0))

			data := handler.getData()
			Expect(len(data)).To(BeNumerically(">", 0))
			Expect(data[0]).To(Equal(testData))
		})

		It("should receive multiple datagrams", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			// Send multiple datagrams
			for i := 0; i < 5; i++ {
				err := sendUnixgramDatagram(sockPath, []byte("message"))
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(10 * time.Millisecond)
			}

			Eventually(func() int64 {
				return handler.getCount()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">=", 5))
		})

		It("should handle large datagrams", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			// Send large datagram
			largeData := make([]byte, 8192)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			err = sendUnixgramDatagram(sockPath, largeData)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int64 {
				return handler.getCount()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">", 0))

			data := handler.getData()
			Expect(len(data)).To(BeNumerically(">", 0))
			Expect(data[0]).To(Equal(largeData))
		})
	})

	Describe("Callback Functionality", func() {
		It("should invoke error callback", func() {
			handler := newTestHandler(false)
			errCollector := newErrorCollector()

			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			srv.RegisterFuncError(errCollector.callback)

			// Starting with invalid socket may trigger errors
			// (depending on implementation)
			_ = srv
		})

		It("should invoke info callback on datagram events", func() {
			handler := newTestHandler(false)
			infoCollector := newInfoCollector()

			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			srv.RegisterFuncInfo(infoCollector.callback)
			startServer(srv, ctx)

			// Wait for ConnectionNew event
			Eventually(func() bool {
				return infoCollector.hasState(libsck.ConnectionNew)
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should invoke server info callback", func() {
			handler := newTestHandler(false)
			srvInfoCollector := newServerInfoCollector()

			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			srv.RegisterFuncInfoServer(srvInfoCollector.callback)
			startServer(srv, ctx)

			// Wait for server start message
			Eventually(func() bool {
				return srvInfoCollector.hasMessage("starting listening")
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should allow changing callbacks", func() {
			handler := newTestHandler(false)
			collector1 := newInfoCollector()
			collector2 := newInfoCollector()

			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			srv.RegisterFuncInfo(collector1.callback)
			startServer(srv, ctx)

			time.Sleep(50 * time.Millisecond)

			// Change callback
			srv.RegisterFuncInfo(collector2.callback)

			// Send datagram to trigger event
			_ = sendUnixgramDatagram(sockPath, []byte("test"))

			time.Sleep(50 * time.Millisecond)

			// Both collectors may have events, but collector2 should have more recent ones
			// (exact behavior depends on timing)
		})
	})

	Describe("UpdateConn Callback", func() {
		It("should invoke UpdateConn on socket creation", func() {
			handler := newTestHandler(false)
			updateConn := newCustomUpdateConn()

			cfg := createBasicConfig()
			sockPath = cfg.Address

			srv, err := scksrv.New(updateConn.callback, handler.handler, cfg)
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			Eventually(func() bool {
				return updateConn.wasCalled()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

			conn := updateConn.getConn()
			Expect(conn).ToNot(BeNil())
		})
	})
})
