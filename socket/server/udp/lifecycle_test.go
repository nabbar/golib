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

package udp_test

import (
	"context"
	"time"

	libsck "github.com/nabbar/golib/socket"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Server Lifecycle", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		srv     libsck.Server
		address string
	)
	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(x, 30*time.Second)
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
		It("should start successfully", func() {
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})
	Describe("Shutdown", func() {
		It("should stop server gracefully", func() {
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
			Expect(srv.Shutdown(ctx)).ToNot(HaveOccurred())
			waitForServerStopped(srv, 5*time.Second)
		})
	})
})
