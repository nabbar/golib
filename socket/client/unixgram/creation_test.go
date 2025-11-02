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
	sckclt "github.com/nabbar/golib/socket/client/unixgram"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UNIX Datagram Client Creation", func() {
	Describe("New", func() {
		Context("with valid socket paths", func() {
			It("should create client with absolute path", func() {
				cli := sckclt.New("/tmp/test.sock")
				Expect(cli).ToNot(BeNil())
			})

			It("should create client with relative path", func() {
				cli := sckclt.New("./test.sock")
				Expect(cli).ToNot(BeNil())
			})

			It("should create client with long path", func() {
				longPath := "/tmp/very/long/path/to/unixgram/socket/test.sock"
				cli := sckclt.New(longPath)
				Expect(cli).ToNot(BeNil())
			})

			It("should create client with special characters in path", func() {
				cli := sckclt.New("/tmp/test-socket_123.sock")
				Expect(cli).ToNot(BeNil())
			})
		})

		Context("with invalid socket paths", func() {
			It("should return nil for empty path", func() {
				cli := sckclt.New("")
				Expect(cli).To(BeNil())
			})
		})

		Context("edge cases", func() {
			It("should accept path with dots", func() {
				cli := sckclt.New("../test.sock")
				Expect(cli).ToNot(BeNil())
			})

			It("should accept path without extension", func() {
				cli := sckclt.New("/tmp/mysocket")
				Expect(cli).ToNot(BeNil())
			})
		})
	})

	Describe("Initial State", func() {
		It("should not be connected initially", func() {
			socketPath := getTestSocketPath()
			cli := createClient(socketPath)

			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should have stable state across multiple checks", func() {
			socketPath := getTestSocketPath()
			cli := createClient(socketPath)

			for i := 0; i < 10; i++ {
				Expect(cli.IsConnected()).To(BeFalse())
			}
		})
	})

	Describe("Multiple Clients", func() {
		It("should create independent clients to same socket", func() {
			socketPath := getTestSocketPath()

			cli1 := createClient(socketPath)
			cli2 := createClient(socketPath)

			Expect(cli1).ToNot(BeNil())
			Expect(cli2).ToNot(BeNil())
			Expect(cli1).ToNot(Equal(cli2))
		})

		It("should create independent clients to different sockets", func() {
			socketPath1 := getTestSocketPath()
			socketPath2 := getTestSocketPath()

			cli1 := createClient(socketPath1)
			cli2 := createClient(socketPath2)

			Expect(cli1).ToNot(BeNil())
			Expect(cli2).ToNot(BeNil())
			Expect(cli1).ToNot(Equal(cli2))
		})
	})
})
