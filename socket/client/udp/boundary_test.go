/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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
	"bytes"
	"net"

	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Client Boundary Tests", func() {
	Describe("Datagram Size Limits", func() {
		It("should handle minimum size datagram (1 byte)", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			data := []byte{0x42}
			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(1))
		})

		It("should handle empty datagram (0 bytes)", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			data := []byte{}
			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should handle safe MTU size (1400 bytes)", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			// Safe size below Ethernet MTU
			data := make([]byte, 1400)
			for i := range data {
				data[i] = byte(i % 256)
			}

			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(1400))
		})

		It("should handle standard Ethernet MTU (1472 bytes)", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			// Max UDP payload for standard Ethernet without fragmentation
			data := make([]byte, 1472)
			for i := range data {
				data[i] = byte(i % 256)
			}

			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(1472))
		})

		It("should handle large datagram (8192 bytes)", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			// Larger datagram (may fragment)
			data := make([]byte, 8192)
			for i := range data {
				data[i] = byte(i % 256)
			}

			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(8192))
		})
	})

	Describe("Address Boundary Cases", func() {
		It("should handle port 0 (dynamic port)", func() {
			cli, err := sckclt.New("127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
		})

		It("should handle port 1 (minimum valid port)", func() {
			cli, err := sckclt.New("127.0.0.1:1")
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
		})

		It("should handle port 65535 (maximum valid port)", func() {
			cli, err := sckclt.New("127.0.0.1:65535")
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
		})

		It("should reject port 65536 (over maximum)", func() {
			cli, err := sckclt.New("127.0.0.1:65536")
			Expect(err).To(HaveOccurred())
			Expect(cli).To(BeNil())
		})

		It("should handle minimum IPv4 address", func() {
			cli, err := sckclt.New("0.0.0.0:8080")
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
		})

		It("should handle maximum IPv4 address", func() {
			cli, err := sckclt.New("255.255.255.255:8080")
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
		})

		It("should handle IPv6 loopback", func() {
			cli, err := sckclt.New("[::1]:8080")
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
		})

		It("should handle IPv6 any address", func() {
			cli, err := sckclt.New("[::]:8080")
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
		})
	})

	Describe("Data Content Boundary Cases", func() {
		It("should handle all-zero datagram", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			data := make([]byte, 100)
			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(100))
		})

		It("should handle all-ones datagram", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			data := make([]byte, 100)
			for i := range data {
				data[i] = 0xFF
			}

			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(100))
		})

		It("should handle binary data with null bytes", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			data := []byte{0x00, 0x01, 0x00, 0x02, 0x00}
			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
		})

		It("should handle all possible byte values", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			data := make([]byte, 256)
			for i := 0; i < 256; i++ {
				data[i] = byte(i)
			}

			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(256))
		})
	})

	Describe("Once() Boundary Cases", func() {
		It("should handle empty request in Once", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			request := bytes.NewBuffer(nil)
			err := cli.Once(ctx, request, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle nil response callback in Once", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			request := bytes.NewBufferString("test")
			err := cli.Once(ctx, request, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle large request in Once", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			data := make([]byte, 1400)
			for i := range data {
				data[i] = byte(i % 256)
			}
			request := bytes.NewBuffer(data)

			err := cli.Once(ctx, request, nil)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("State Transition Boundaries", func() {
		It("should handle multiple connect/close cycles", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			for i := 0; i < 5; i++ {
				err := cli.Connect(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(cli.IsConnected()).To(BeTrue())

				err = cli.Close()
				Expect(err).ToNot(HaveOccurred())
				Expect(cli.IsConnected()).To(BeFalse())
			}
		})

		It("should handle reconnect without explicit close", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			// First connect
			err := cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Reconnect without close (should replace socket)
			err = cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeTrue())
		})
	})

	Describe("Callback Registration Boundaries", func() {
		It("should handle nil then valid error callback", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			cli.RegisterFuncError(nil)
			cli.RegisterFuncError(func(errs ...error) {})

			Expect(cli).ToNot(BeNil())
		})

		It("should handle valid then nil error callback", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			cli.RegisterFuncError(func(errs ...error) {})
			cli.RegisterFuncError(nil)

			Expect(cli).ToNot(BeNil())
		})

		It("should handle nil then valid info callback", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			cli.RegisterFuncInfo(nil)
			cli.RegisterFuncInfo(func(_, _ net.Addr, _ libsck.ConnState) {})

			Expect(cli).ToNot(BeNil())
		})
	})
})
