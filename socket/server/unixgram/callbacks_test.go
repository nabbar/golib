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
	"net"
	"os"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/unixgram"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unixgram Callbacks", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		path   string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(x, 30*time.Second)
		path = getTempSocketPath()
	})

	AfterEach(func() {
		_ = os.Remove(path)

		if cancel != nil {
			cancel()
		}
	})

	Describe("RegisterFuncError", func() {
		It("should register", func() {
			var cnt atomic.Int64
			srv := scksrv.New(nil, echoHandler)

			srv.RegisterFuncError(func(errs ...error) {
				for range errs {
					cnt.Add(1)
				}
			})

			Expect(srv.RegisterSocket(path, 0600, -1)).ToNot(HaveOccurred())
		})
	})
	Describe("RegisterFuncInfo", func() {
		It("should invoke on datagram", func() {
			var cnt atomic.Int64

			srv := scksrv.New(nil, echoHandler)
			defer func() {
				_ = srv.Shutdown(ctx)
			}()

			srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				cnt.Add(1)
			})

			Expect(srv.RegisterSocket(path, 0600, -1)).ToNot(HaveOccurred())

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			Expect(sendDatagram(path, []byte("test"))).ToNot(HaveOccurred())

			time.Sleep(200 * time.Millisecond)
			Expect(srv.Shutdown(ctx)).ToNot(HaveOccurred())

			time.Sleep(200 * time.Millisecond)
			Expect(cnt.Load()).To(BeNumerically(">", 0))
		})
	})
	Describe("RegisterFuncInfoServer", func() {
		It("should invoke", func() {
			var cnt atomic.Int64

			srv := scksrv.New(nil, echoHandler)
			defer func() {
				_ = srv.Shutdown(ctx)
			}()

			srv.RegisterFuncInfoServer(func(msg string) {
				cnt.Add(1)
			})

			Expect(srv.RegisterSocket(path, 0600, -1)).ToNot(HaveOccurred())

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			Expect(srv.Shutdown(ctx)).ToNot(HaveOccurred())
			time.Sleep(200 * time.Millisecond)

			Expect(cnt.Load()).To(BeNumerically(">", 0))
		})
	})
})
