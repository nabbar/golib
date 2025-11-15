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
	sckclt "github.com/nabbar/golib/socket/client/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Client Creation", func() {
	Describe("New", func() {
		Context("with valid addresses", func() {
			It("should create a new client with localhost and port", func() {
				cli, err := sckclt.New("127.0.0.1:8080")
				Expect(err).ToNot(HaveOccurred())
				Expect(cli).ToNot(BeNil())
			})

			It("should create a new client with 0.0.0.0 and port", func() {
				cli, err := sckclt.New("0.0.0.0:8081")
				Expect(err).ToNot(HaveOccurred())
				Expect(cli).ToNot(BeNil())
			})

			It("should create a new client with only port", func() {
				cli, err := sckclt.New(":8082")
				Expect(err).ToNot(HaveOccurred())
				Expect(cli).ToNot(BeNil())
			})

			It("should create a new client with hostname", func() {
				cli, err := sckclt.New("localhost:8083")
				Expect(err).ToNot(HaveOccurred())
				Expect(cli).ToNot(BeNil())
			})

			It("should create multiple independent clients", func() {
				cli1, err1 := sckclt.New("127.0.0.1:8084")
				cli2, err2 := sckclt.New("127.0.0.1:8085")

				Expect(err1).ToNot(HaveOccurred())
				Expect(err2).ToNot(HaveOccurred())
				Expect(cli1).ToNot(BeNil())
				Expect(cli2).ToNot(BeNil())
				Expect(cli1).ToNot(Equal(cli2))
			})
		})

		Context("with invalid addresses", func() {
			It("should fail with empty address", func() {
				cli, err := sckclt.New("")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(sckclt.ErrAddress))
				Expect(cli).To(BeNil())
			})

			It("should fail with malformed address", func() {
				cli, err := sckclt.New("not-a-valid-address")
				Expect(err).To(HaveOccurred())
				Expect(cli).To(BeNil())
			})

			It("should fail with port only without colon", func() {
				cli, err := sckclt.New("8080")
				Expect(err).To(HaveOccurred())
				Expect(cli).To(BeNil())
			})

			It("should fail with invalid port", func() {
				cli, err := sckclt.New("127.0.0.1:99999")
				Expect(err).To(HaveOccurred())
				Expect(cli).To(BeNil())
			})

			It("should fail with invalid characters in port", func() {
				cli, err := sckclt.New("127.0.0.1:abc")
				Expect(err).To(HaveOccurred())
				Expect(cli).To(BeNil())
			})
		})

		Context("with edge case addresses", func() {
			It("should handle IPv6 localhost", func() {
				cli, err := sckclt.New("[::1]:8086")
				Expect(err).ToNot(HaveOccurred())
				Expect(cli).ToNot(BeNil())
			})

			It("should handle IPv6 address", func() {
				cli, err := sckclt.New("[2001:db8::1]:8087")
				Expect(err).ToNot(HaveOccurred())
				Expect(cli).ToNot(BeNil())
			})

			It("should handle port 0 (dynamic port)", func() {
				cli, err := sckclt.New("127.0.0.1:0")
				Expect(err).ToNot(HaveOccurred())
				Expect(cli).ToNot(BeNil())
			})
		})
	})

	Describe("Initial State", func() {
		var cli sckclt.ClientTCP

		BeforeEach(func() {
			cli = createClient(getTestAddress())
		})

		It("should not be connected initially", func() {
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should be safe to call IsConnected multiple times", func() {
			for i := 0; i < 10; i++ {
				Expect(cli.IsConnected()).To(BeFalse())
			}
		})
	})
})
