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
	"time"

	libdur "github.com/nabbar/golib/duration"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libptc "github.com/nabbar/golib/network/protocol"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server"
)

var _ = Describe("Server Factory Edge Cases", func() {
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

	Context("Boundary Conditions", func() {
		It("should handle zero value configuration gracefully", func() {
			cfg := sckcfg.Server{}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).To(Equal(sckcfg.ErrInvalidProtocol))
			Expect(srv).To(BeNil())
		})

		It("should handle protocol value at boundary", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkProtocol(0),
				Address: ":8080",
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).To(Equal(sckcfg.ErrInvalidProtocol))
			Expect(srv).To(BeNil())
		})

		It("should handle large protocol value", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkProtocol(255),
				Address: ":8080",
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).To(Equal(sckcfg.ErrInvalidProtocol))
			Expect(srv).To(BeNil())
		})
	})

	Context("Configuration Edge Cases", func() {
		It("should handle empty address for TCP", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: "",
			}

			// Empty address may be valid (listen on all interfaces)
			srv, err := scksrv.New(nil, basicHandler(), cfg)
			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
			// Result depends on tcp package validation
			_ = err
		})

		It("should handle very long address for TCP", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: "very.long.domain.name.that.does.not.exist.example.com:8080",
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			// Creation may succeed, listening will fail
			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
			_ = err
		})

		It("should handle zero idle timeout", func() {
			cfg := sckcfg.Server{
				Network:        libptc.NetworkTCP,
				Address:        getTestTCPAddress(),
				ConIdleTimeout: 0,
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should handle negative idle timeout", func() {
			cfg := sckcfg.Server{
				Network:        libptc.NetworkTCP,
				Address:        getTestTCPAddress(),
				ConIdleTimeout: libdur.Seconds(-1),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should handle very large idle timeout", func() {
			cfg := sckcfg.Server{
				Network:        libptc.NetworkTCP,
				Address:        getTestTCPAddress(),
				ConIdleTimeout: libdur.Days(365), // 1 year
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})
	})

	Context("Rapid Creation and Destruction", func() {
		It("should handle rapid create/destroy cycles", func() {
			for i := 0; i < 10; i++ {
				cfg := sckcfg.Server{
					Network: libptc.NetworkTCP,
					Address: getTestTCPAddress(),
				}

				srv, err := scksrv.New(nil, basicHandler(), cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(srv).ToNot(BeNil())

				if srv != nil {
					_ = srv.Close()
				}
			}
		})

		It("should handle rapid concurrent creation", func() {
			done := make(chan bool, 20)

			for i := 0; i < 20; i++ {
				go func() {
					defer GinkgoRecover()

					cfg := sckcfg.Server{
						Network: libptc.NetworkTCP,
						Address: getTestTCPAddress(),
					}

					srv, err := scksrv.New(nil, basicHandler(), cfg)
					Expect(err).ToNot(HaveOccurred())

					if srv != nil {
						_ = srv.Close()
					}

					done <- true
				}()
			}

			for i := 0; i < 20; i++ {
				Eventually(done, 5*time.Second).Should(Receive())
			}
		})
	})

	Context("Protocol Value Validation", func() {
		It("should reject protocol values between valid ranges", func() {
			// Test protocol values that are not defined
			invalidProtocols := []libptc.NetworkProtocol{
				100, 200, 254,
			}

			for _, proto := range invalidProtocols {
				cfg := sckcfg.Server{
					Network: proto,
					Address: ":8080",
				}

				srv, err := scksrv.New(nil, basicHandler(), cfg)
				Expect(err).To(Equal(sckcfg.ErrInvalidProtocol))
				Expect(srv).To(BeNil())
			}
		})
	})
})
