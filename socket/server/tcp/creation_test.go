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

package tcp_test

import (
	"net"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Creation", func() {
	Describe("New", func() {
		Context("with valid parameters", func() {
			It("should create a new server with handler", func() {
				srv := scksrv.New(nil, echoHandler)
				Expect(srv).ToNot(BeNil())
			})

			It("should create a new server with handler and update conn", func() {
				updateConnCalled := false
				updateConn := func(conn net.Conn) {
					updateConnCalled = true
				}

				srv := scksrv.New(updateConn, echoHandler)
				Expect(srv).ToNot(BeNil())

				// updateConn is not called until a connection is made
				Expect(updateConnCalled).To(BeFalse())
			})

			It("should create a new server with nil update conn", func() {
				srv := scksrv.New(nil, echoHandler)
				Expect(srv).ToNot(BeNil())
			})

			It("should create multiple independent servers", func() {
				srv1 := scksrv.New(nil, echoHandler)
				srv2 := scksrv.New(nil, echoHandler)

				Expect(srv1).ToNot(BeNil())
				Expect(srv2).ToNot(BeNil())
				Expect(srv1).ToNot(Equal(srv2))
			})
		})

		Context("with nil handler", func() {
			It("should still create a server but fail on Listen", func() {
				srv := scksrv.New(nil, nil)
				Expect(srv).ToNot(BeNil())
				// Handler validation happens during Listen
			})
		})
	})

	Describe("RegisterServer", func() {
		var srv scksrv.ServerTcp

		BeforeEach(func() {
			srv = scksrv.New(nil, echoHandler)
		})

		Context("with valid addresses", func() {
			It("should register with localhost and port", func() {
				err := srv.RegisterServer("127.0.0.1:8080")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should register with 0.0.0.0 and port", func() {
				err := srv.RegisterServer("0.0.0.0:8081")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should register with IPv6 loopback", func() {
				err := srv.RegisterServer("[::1]:8082")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should register with any available port", func() {
				address := getTestAddress()
				err := srv.RegisterServer(address)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should allow re-registration with different address", func() {
				err := srv.RegisterServer("127.0.0.1:8083")
				Expect(err).ToNot(HaveOccurred())

				err = srv.RegisterServer("127.0.0.1:8084")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with invalid addresses", func() {
			It("should fail with empty address", func() {
				err := srv.RegisterServer("")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(scksrv.ErrInvalidAddress))
			})

			It("should fail with invalid format", func() {
				err := srv.RegisterServer("not-a-valid-address")
				Expect(err).To(HaveOccurred())
			})

			It("should fail with missing port", func() {
				err := srv.RegisterServer("127.0.0.1")
				Expect(err).To(HaveOccurred())
			})

			It("should fail with invalid port", func() {
				err := srv.RegisterServer("127.0.0.1:99999")
				Expect(err).To(HaveOccurred())
			})

			It("should fail with invalid hostname", func() {
				err := srv.RegisterServer("invalid..host:8080")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Initial State", func() {
		var srv libsck.Server

		BeforeEach(func() {
			srv = scksrv.New(nil, echoHandler)
		})

		It("should not be running initially", func() {
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should not be gone initially", func() {
			Expect(srv.IsGone()).To(BeFalse())
		})

		It("should have no open connections initially", func() {
			Expect(srv.OpenConnections()).To(Equal(int64(0)))
		})

		It("should have a Done channel", func() {
			done := srv.Done()
			Expect(done).ToNot(BeNil())
			// Channel exists and can be used for signaling
		})
	})
})
