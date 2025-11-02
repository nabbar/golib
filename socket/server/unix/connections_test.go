//go:build linux || darwin

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

package unix_test

import (
	"context"
	"os"
	"time"

	libsck "github.com/nabbar/golib/socket"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unix Socket Connections", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		srv    libsck.Server
		path   string
	)
	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(x, 30*time.Second)

		path = getTempSocketPath()

		srv = createAndRegisterServer(path, echoHandler)
		startServer(ctx, srv)

		waitForServerRunning(srv, 2*time.Second)
	})
	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}

		_ = os.Remove(path)

		if cancel != nil {
			cancel()
		}
	})
	Describe("Single Connection", func() {
		It("should accept connection", func() {
			conn, err := connectUnixClient(path)
			defer func() {
				_ = conn.Close()
			}()

			Expect(err).ToNot(HaveOccurred())
			waitForConnections(srv, 1, 2*time.Second)
		})
		It("should echo data", func() {
			conn, _ := connectUnixClient(path)
			defer func() {
				_ = conn.Close()
			}()

			msg := []byte("test")
			_, e := conn.Write(msg)
			Expect(e).ToNot(HaveOccurred())

			buf := make([]byte, 10)
			n, e := conn.Read(buf)
			Expect(e).ToNot(HaveOccurred())
			Expect(buf[:n]).To(Equal(msg))
		})
	})
	Describe("Multiple Connections", func() {
		It("should handle multiple clients", func() {
			conns := make([]interface{}, 3)

			for i := 0; i < 3; i++ {
				c, _ := connectUnixClient(path)

				defer func() {
					_ = c.Close()
				}()

				conns[i] = c
			}

			time.Sleep(200 * time.Millisecond)
			Expect(srv.OpenConnections()).To(BeNumerically(">=", 1))
		})
	})
	Describe("Connection Cleanup", func() {
		It("should cleanup after disconnect", func() {
			conn, _ := connectUnixClient(path)
			defer func() {
				_ = conn.Close()
			}()
			waitForConnections(srv, 1, 2*time.Second)

			_ = srv.Close()
			waitForConnections(srv, 0, 5*time.Second)
		})
	})
})
