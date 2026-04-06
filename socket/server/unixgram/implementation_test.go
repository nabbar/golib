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
	"io"
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

		It("should cover context methods within handler", func() {
			done := make(chan struct{})
			handler := func(c libsck.Context) {
				time.Sleep(time.Second)

				defer close(done)
				// Coverage for context methods
				_, _ = c.Deadline()
				_ = c.Done()
				_ = c.Err()
				_ = c.Value("key")
				Expect(c.IsConnected()).To(BeTrue())
				Expect(c.LocalHost()).To(ContainSubstring("unixgram"))
				Expect(c.RemoteHost()).To(BeEmpty())

				// Write is disabled in unixgram
				_, err := c.Write([]byte("test"))
				Expect(err).To(HaveOccurred())
			}

			srv, path, _ := createServerWithHandler(handler)
			sockPath = path
			startServer(srv, ctx)
			defer srv.Close()

			err := sendUnixgramDatagram(sockPath, []byte("trigger"))
			Expect(err).ToNot(HaveOccurred())
			Eventually(done, 5*time.Second).Should(BeClosed())
		})

		It("should handle context Close, double Close and post-close state", func() {
			var capturedCtx libsck.Context
			done := make(chan struct{})
			keepAlive := make(chan struct{})

			handler := func(c libsck.Context) {
				capturedCtx = c
				close(done)
				// Keep the handler goroutine alive to keep the server in "Running" state
				// while the test performs assertions on the captured context.
				<-keepAlive
			}

			srv, path, _ := createServerWithHandler(handler)
			sockPath = path
			startServer(srv, ctx)
			defer srv.Close()

			_ = sendUnixgramDatagram(sockPath, []byte("trigger"))
			Eventually(done, 2*time.Second).Should(BeClosed())

			if capturedCtx != nil {
				// First Close
				Expect(capturedCtx.Close()).To(Succeed())
				// Double Close
				Expect(capturedCtx.Close()).To(Succeed())

				// Methods after close
				Expect(capturedCtx.IsConnected()).To(BeFalse())
				_, err := capturedCtx.Read(make([]byte, 10))
				Expect(err).To(Equal(io.ErrClosedPipe))

				_, err = capturedCtx.Write(make([]byte, 10))
				Expect(err).To(Equal(io.ErrClosedPipe))

				// Target context.go coverage for edge cases
				Expect(capturedCtx.Value(nil)).To(BeNil())
				Expect(capturedCtx.Err()).To(Equal(io.ErrClosedPipe))
				Expect(capturedCtx.LocalHost()).To(ContainSubstring("unixgram"))
				_ = capturedCtx.RemoteHost()
			}
			close(keepAlive)
		})

		It("should handle read errors and onCloseClose coverage", func() {
			done := make(chan struct{})
			srv, path, _ := createServerWithHandler(func(c libsck.Context) {
				time.Sleep(time.Second)

				defer close(done)
				// Close the connection underneath to trigger error in Read
				_ = c.Close()
				_, err := c.Read(make([]byte, 10))
				Expect(err).To(Equal(io.ErrClosedPipe))
			})
			sockPath = path
			startServer(srv, ctx)
			defer srv.Close()

			_ = sendUnixgramDatagram(sockPath, []byte("trigger"))
			Eventually(done, 5*time.Second).Should(BeClosed())
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

			log, ok := srv.(srvExport)
			Expect(ok).To(BeTrue())

			// Trigger fctError coverage with multiple errors
			log.TestFctError(io.EOF, io.ErrUnexpectedEOF)
			Expect(len(errCollector.getErrors())).To(Equal(2))

			// Test callback removal
			srv.RegisterFuncError(nil)
			errCollector.clear()
			log.TestFctError(io.EOF)
			Expect(errCollector.hasErrors()).To(BeFalse())
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

			// Test removal
			srv.RegisterFuncInfo(nil)
			infoCollector.clear()
			log, ok := srv.(srvExport)
			Expect(ok).To(BeTrue())
			log.TestFctInfo(nil, nil, libsck.ConnectionNew)
			Expect(infoCollector.hasState(libsck.ConnectionNew)).To(BeFalse())
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

			log, ok := srv.(srvExport)
			Expect(ok).To(BeTrue())

			// Test arguments and removal
			log.TestFctInfoSrv("test message %s", "arg")
			Expect(srvInfoCollector.hasMessage("test message arg")).To(BeTrue())

			srv.RegisterFuncInfoServer(nil)
			srvInfoCollector.clear()
			log.TestFctInfoSrv("hidden")
			Expect(srvInfoCollector.hasMessage("hidden")).To(BeFalse())
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

	Describe("Internal State and Resource Management", func() {
		It("should cover internal file and permission helpers", func() {
			srv, path, err := createServerWithHandler(echoHandler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			log, ok := srv.(srvExport)
			Expect(ok).To(BeTrue())

			// Cover internal helpers in listener.go and model.go
			_, _ = log.TestGetSocketFile()
			_ = log.TestGetSocketPerm()
			_ = log.TestGetSocketGroup()
			_, _ = log.TestCheckFile(path)
			_ = log.TestGetGoneChan()

			// Context and Pool coverage (model.go and context.go)
			c := log.TestGetContext(ctx, nil, nil, "loc")
			Expect(c).ToNot(BeNil())

			// Additional context logic coverage
			_, _ = c.Deadline()
			_ = c.Done()
			Expect(c.Err()).To(BeNil())

			log.TestPutContext(c)
		})

		It("should cover nil value for internal struct", func() {
			// Nil Server methods (model.go guards)
			nilSrv := scksrv.NewTestNilServer()
			if ls, ok := nilSrv.(srvExport); ok {
				ls.TestFctError(io.EOF)
				ls.TestFctInfo(nil, nil, libsck.ConnectionNew)
				ls.TestFctInfoSrv("test")
			}

			nilSrv.RegisterFuncError(nil)
			nilSrv.RegisterFuncInfo(nil)
			nilSrv.RegisterFuncInfoServer(nil)
			_ = nilSrv.OpenConnections()
		})

		It("should cover nil context manipulation for internal context", func() {
			// Empty Context methods (context.go guards)
			emptyCtx := scksrv.NewTestEmptyContext()
			_, _ = emptyCtx.Deadline()
			Expect(emptyCtx.Done()).ToNot(BeNil())
			Expect(emptyCtx.Value("key")).To(BeNil())
			_ = emptyCtx.Close()
			Expect(emptyCtx.Err()).ToNot(BeNil())
		})

		It("should cover nil io.read/Writer for internal context", func() {
			// Empty Context methods (context.go guards)
			emptyCtx := scksrv.NewTestEmptyContext()
			_, _ = emptyCtx.Read(nil)
			_, _ = emptyCtx.Write(nil)

			// Specific branch in onErrorClose
			if ec, ok := emptyCtx.(ctxExport); ok {
				_ = ec.TestOnErrorClose(nil)
				_ = ec.TestOnErrorClose(io.EOF)
			}
		})
	})
})
