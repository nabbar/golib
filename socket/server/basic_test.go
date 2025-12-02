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

package server_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server"
)

var _ = Describe("Server Factory Basic Operations", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
		time.Sleep(50 * time.Millisecond)
	})

	Context("Server Interface Implementation", func() {
		It("should return server implementing socket.Server interface", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			// Verify interface implementation
			var _ libsck.Server = srv

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should allow calling Listen on created server", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			// Start server in background
			go func() {
				_ = srv.Listen(ctx)
			}()

			time.Sleep(100 * time.Millisecond)

			if srv != nil {
				Expect(srv.IsRunning()).To(BeTrue())
				_ = srv.Shutdown(ctx)
			}
		})

		It("should allow calling Shutdown on created server", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			go func() {
				_ = srv.Listen(ctx)
			}()

			time.Sleep(100 * time.Millisecond)

			if srv != nil {
				err := srv.Shutdown(ctx)
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("should allow checking IsRunning on created server", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				Expect(srv.IsRunning()).To(BeFalse())
				_ = srv.Shutdown(ctx)
			}
		})

		It("should allow counting OpenConnections on created server", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				count := srv.OpenConnections()
				Expect(count).To(BeNumerically(">=", 0))
				_ = srv.Shutdown(ctx)
			}
		})
	})

	Context("UpdateConn Callback", func() {
		It("should accept nil UpdateConn callback", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should accept custom UpdateConn callback", func() {
			upd := func(c net.Conn) {
				_, _ = fmt.Fprintf(io.Discard, "%s->%s", c.RemoteAddr().String(), c.LocalAddr().String())
			}

			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(upd, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})
	})

	Context("Handler Function", func() {
		It("should accept basic handler", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should accept echo handler", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, echoHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})
	})

	Context("Protocol Delegation", func() {
		It("should delegate TCP to tcp package", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should delegate UDP to udp package", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkUDP,
				Address: getTestUDPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})
	})
})
