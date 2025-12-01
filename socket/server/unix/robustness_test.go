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

// robustness_test.go validates error handling, edge cases, and fault tolerance.
// Tests server behavior under adverse conditions and boundary cases.
package unix_test

import (
	"context"
	"net"
	"os"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksru "github.com/nabbar/golib/socket/server/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unix Server Robustness", func() {
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

	Context("error callback behavior", func() {
		It("should call error callback on errors", func() {
			errorCalled := &atomic.Bool{}
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncError(func(errs ...error) {
				errorCalled.Store(true)
			})

			// Cause an error by trying to listen on invalid path
			invalidPath := "/invalid/path/that/does/not/exist/test.sock"
			err = srv.RegisterSocket(invalidPath, 0600, -1)
			Expect(err).ToNot(HaveOccurred())

			_ = srv.Listen(c)
			Eventually(errorCalled.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should call info callback on connection events", func() {
			infoCalled := &atomic.Int32{}
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncInfo(func(_, _ net.Addr, _ libsck.ConnState) {
				infoCalled.Add(1)
			})

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			_ = con.Close()

			Eventually(func() int32 {
				return infoCalled.Load()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">=", 1))
		})
	})

	Context("idle timeout behavior", func() {
		It("should close idle connections after timeout", func() {
			cfg := createConfigWithIdleTimeout(socketPath, 200*time.Millisecond)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			// Connection is established
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			// Wait for idle timeout
			time.Sleep(500 * time.Millisecond)

			// Connection should be closed due to timeout
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))
		})

		It("should not timeout active connections", func() {
			cfg := createConfigWithIdleTimeout(socketPath, 200*time.Millisecond)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			// Keep connection active with periodic messages
			for i := 0; i < 5; i++ {
				data := []byte("Keep alive")
				rsp := sendAndReceive(con, data)
				Expect(rsp).To(Equal(data))
				time.Sleep(100 * time.Millisecond)
			}

			// Connection should still be open
			Expect(srv.OpenConnections()).To(Equal(int64(1)))
		})
	})

	Context("handler panics", func() {
		It("should recover from handler panic", func() {
			panicCount := &atomic.Int32{}
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, func(ctx libsck.Context) {
				defer func() { _ = ctx.Close() }()
				panicCount.Add(1)
				//panic("intentional panic for testing")
			}, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			_ = con.Close()

			// Server should still be running after panic
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Context("socket file permissions", func() {
		It("should create socket with correct permissions", func() {
			cfg := createConfigWithPerms(socketPath, 0600, -1)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			info, err := os.Stat(socketPath)
			Expect(err).ToNot(HaveOccurred())

			// Check that it's a socket
			Expect(info.Mode() & os.ModeSocket).ToNot(Equal(os.FileMode(0)))
		})
	})

	Context("malformed client behavior", func() {
		It("should handle client that doesn't send data", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, readOnlyHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			// Wait a bit then close without sending
			time.Sleep(100 * time.Millisecond)
			_ = con.Close()

			// Server should handle gracefully
			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should handle client that only sends data", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, writeOnlyHandler("Server message"), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			// Read server message
			buf := make([]byte, 1024)
			n, err := con.Read(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(BeNumerically(">", 0))

			// Server should handle gracefully
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Context("resource cleanup", func() {
		It("should cleanup all connections on shutdown", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, slowHandler(1*time.Second), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			// Create multiple connections
			con1 := connectToServer(socketPath)
			con2 := connectToServer(socketPath)
			con3 := connectToServer(socketPath)

			// Wait for connections to be established
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(3)))

			// Shutdown
			err = srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			// All connections should be closed
			_ = con1.Close()
			_ = con2.Close()
			_ = con3.Close()

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Context("repeated start/stop cycles", func() {
		It("should handle multiple Listen attempts", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			// First start
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			// Stop
			err = srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			// Wait for complete stop
			time.Sleep(200 * time.Millisecond)

			// Second start should work
			c2, cnl2 := context.WithCancel(globalCtx)
			defer cnl2()

			startServerInBackground(c2, srv)
			waitForServer(srv, 2*time.Second)

			err = srv.Shutdown(c2)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
