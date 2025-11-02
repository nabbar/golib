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
	scksrv "github.com/nabbar/golib/socket/server/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Server Creation", func() {
	Describe("New", func() {
		It("should create server with valid handler", func() {
			srv := scksrv.New(nil, echoHandler)
			Expect(srv).ToNot(BeNil())
			Expect(srv.IsRunning()).To(BeFalse())
		})
	})

	Describe("RegisterServer", func() {
		It("should accept valid address", func() {
			srv := scksrv.New(nil, echoHandler)
			err := srv.RegisterServer("127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reject empty address", func() {
			srv := scksrv.New(nil, echoHandler)
			err := srv.RegisterServer("")
			Expect(err).To(HaveOccurred())
		})

		It("should reject invalid address", func() {
			srv := scksrv.New(nil, echoHandler)
			err := srv.RegisterServer("invalid:address:format")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Initial State", func() {
		It("should not be running initially", func() {
			srv := scksrv.New(nil, echoHandler)
			_ = srv.RegisterServer(getTestAddress())
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should be gone initially", func() {
			srv := scksrv.New(nil, echoHandler)
			_ = srv.RegisterServer(getTestAddress())
			Expect(srv.IsGone()).To(BeTrue())
		})

		It("should have zero connections initially", func() {
			srv := scksrv.New(nil, echoHandler)
			_ = srv.RegisterServer(getTestAddress())
			Expect(srv.OpenConnections()).To(Equal(int64(0)))
		})
	})
})
