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

	scksrv "github.com/nabbar/golib/socket/server/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Server Error Handling", func() {
	Describe("Invalid Configuration", func() {
		It("should fail without handler", func() {
			srv := scksrv.New(nil, nil)
			_ = srv.RegisterServer(getTestAddress())
			ctx, cancel := context.WithTimeout(x, 5*time.Second)
			defer cancel()
			startServer(ctx, srv)
			time.Sleep(200 * time.Millisecond)
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should fail without address", func() {
			srv := scksrv.New(nil, echoHandler)
			ctx, cancel := context.WithTimeout(x, 5*time.Second)
			defer cancel()
			startServer(ctx, srv)
			time.Sleep(200 * time.Millisecond)
			Expect(srv.IsRunning()).To(BeFalse())
		})
	})

	Describe("Shutdown", func() {
		It("should handle double shutdown", func() {
			srv := scksrv.New(nil, echoHandler)
			_ = srv.RegisterServer(getTestAddress())
			ctx, cancel := context.WithTimeout(x, 10*time.Second)
			defer cancel()
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
			_ = srv.Shutdown(ctx)
			Expect(func() { _ = srv.Shutdown(ctx) }).ToNot(Panic())
		})
	})

	Describe("SetTLS", func() {
		It("should accept SetTLS call (no-op for UDP)", func() {
			srv := scksrv.New(nil, echoHandler)
			Expect(func() { _ = srv.SetTLS(false, nil) }).ToNot(Panic())
		})
	})
})
