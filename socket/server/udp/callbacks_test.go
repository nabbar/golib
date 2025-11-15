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
	"net"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Server Callbacks", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		address string
	)
	BeforeEach(func() { ctx, cancel = context.WithTimeout(x, 30*time.Second); address = getTestAddress() })
	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("RegisterFuncError", func() {
		It("should invoke error callback", func() {
			var errorCount atomic.Int64
			srv := scksrv.New(nil, echoHandler)
			srv.RegisterFuncError(func(errs ...error) {
				for range errs {
					errorCount.Add(1)
				}
			})
			err := srv.RegisterServer(address)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("RegisterFuncInfo", func() {
		It("should invoke info callback on datagram", func() {
			var infoCount atomic.Int64
			srv := scksrv.New(nil, echoHandler)
			srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) { infoCount.Add(1) })
			_ = srv.RegisterServer(address)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
			_ = sendDatagram(address, []byte("test"))
			time.Sleep(300 * time.Millisecond)
			if srv != nil && srv.IsRunning() {
				_ = srv.Shutdown(ctx)
			}
			Expect(infoCount.Load()).To(BeNumerically(">", 0))
		})
	})

	Describe("RegisterFuncInfoServer", func() {
		It("should invoke server info callback", func() {
			var infoCount atomic.Int64
			srv := scksrv.New(nil, echoHandler)
			srv.RegisterFuncInfoServer(func(msg string) { infoCount.Add(1) })
			_ = srv.RegisterServer(address)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
			if srv != nil && srv.IsRunning() {
				_ = srv.Shutdown(ctx)
			}
			Expect(infoCount.Load()).To(BeNumerically(">", 0))
		})
	})
})
