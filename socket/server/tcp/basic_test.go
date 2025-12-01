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

// basic_test.go provides fundamental operational tests for the TCP server.
// It validates core functionality including server lifecycle (start/stop),
// connection handling, and graceful shutdown mechanisms.
package tcp_test

import (
	"context"
	"time"

	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Basic Operations", func() {
	var (
		srv scksrt.ServerTcp
		adr string
		c   context.Context
		cnl context.CancelFunc
	)

	BeforeEach(func() {
		adr = getTestAddr()
		cfg := createDefaultConfig(adr)
		var err error
		srv, err = scksrt.New(nil, echoHandler, cfg)
		Expect(err).ToNot(HaveOccurred())

		c, cnl = context.WithCancel(globalCtx)
		time.Sleep(5 * time.Millisecond)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		time.Sleep(50 * time.Millisecond)
	})

	Context("starting and stopping", func() {
		It("should start server successfully", func() {
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			Expect(srv.IsRunning()).To(BeTrue())
			Expect(srv.IsGone()).To(BeFalse())
		})

		It("should accept connections when running", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			Expect(con).ToNot(BeNil())
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))
		})

		It("should echo messages correctly", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			msg := []byte("Hello, World!")
			rsp := sendAndReceive(con, msg)

			Expect(rsp).To(Equal(msg))
		})

		It("should stop server with context cancellation", func() {
			tctx, tcnl := context.WithCancel(c)
			startServerInBackground(tctx, srv)
			waitForServer(srv, 2*time.Second)

			tcnl()

			waitForServerStopped(srv, 2*time.Second)
			Expect(srv.IsRunning()).To(BeFalse())
		})
	})

	Context("connection management", func() {
		It("should track multiple connections", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con1 := connectToServer(adr)
			defer func() { _ = con1.Close() }()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			con2 := connectToServer(adr)
			defer func() { _ = con2.Close() }()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(2)))

			con3 := connectToServer(adr)
			defer func() { _ = con3.Close() }()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(3)))
		})

		It("should decrement count when connection closes", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			_ = con.Close()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})

		It("should handle rapid connection open/close cycles", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			for i := 0; i < 10; i++ {
				con := connectToServer(adr)
				_ = con.Close()
			}

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Context("graceful shutdown", func() {
		It("should shutdown gracefully with Shutdown()", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			_ = con.Close()

			err := srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			Expect(srv.IsRunning()).To(BeFalse())
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})

		It("should close immediately with Close()", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			err := srv.Close()
			Expect(err).ToNot(HaveOccurred())

			waitForServerStopped(srv, 2*time.Second)
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should set IsGone flag after shutdown", func() {
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			err := srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return srv.IsGone()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})
	})

	Context("error conditions", func() {
		It("should handle context cancellation during Listen", func() {
			tctx, tcnl := context.WithTimeout(c, 500*time.Microsecond)
			defer tcnl()

			err := srv.Listen(tctx)
			// No error expected as context cancellation is normal
			Expect(err).To(HaveOccurred())
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should handle double close gracefully", func() {
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			err1 := srv.Close()
			Expect(err1).ToNot(HaveOccurred())

			// Second close should not panic
			err2 := srv.Close()
			// May return error or nil, but should not panic
			_ = err2
		})
	})
})
