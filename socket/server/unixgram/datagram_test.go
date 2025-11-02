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

package unixgram_test

import (
	"context"
	"os"
	"time"

	libsck "github.com/nabbar/golib/socket"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unixgram Datagram", func() {
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

		time.Sleep(300 * time.Millisecond)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Shutdown(ctx)
		}

		_ = os.Remove(path)

		if cancel != nil {
			cancel()
		}
	})

	Describe("Send Datagram", func() {
		It("should send single", func() {
			Expect(sendDatagram(path, []byte("test"))).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
		})
	})
	Describe("Connection Count", func() {
		It("should show 1 active", func() {
			time.Sleep(100 * time.Millisecond)
			Expect(srv.OpenConnections()).To(Equal(int64(1)))
		})
	})
})
