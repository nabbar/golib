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

package client_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libptc "github.com/nabbar/golib/network/protocol"
	sckcfg "github.com/nabbar/golib/socket/config"
	sckclt "github.com/nabbar/golib/socket/client"
)

var _ = Describe("Basic Client Operations", func() {
	Context("Client Lifecycle", func() {
		It("should create and close TCP client", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:7001",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			err = cli.Close()
			// Close may return error if not connected
			_ = err
		})

		It("should create and close UDP client", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkUDP,
				Address: "localhost:7002",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			err = cli.Close()
			// Close may return error if not connected
			_ = err
		})
	})

	Context("Client Interface Implementation", func() {
		It("should implement socket.Client interface for TCP", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:7003",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
			defer cli.Close()

			// Verify that client has Close method
			var closer interface{ Close() error } = cli
			Expect(closer).ToNot(BeNil())
		})

		It("should implement socket.Client interface for UDP", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkUDP,
				Address: "localhost:7004",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
			defer cli.Close()

			// Verify that client has Close method
			var closer interface{ Close() error } = cli
			Expect(closer).ToNot(BeNil())
		})
	})

	Context("Protocol Validation", func() {
		It("should validate TCP protocol", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:7005",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
			defer cli.Close()
		})

		It("should validate UDP protocol", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkUDP,
				Address: "localhost:7006",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
			defer cli.Close()
		})

		It("should reject invalid protocol", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkProtocol(99),
				Address: "localhost:7007",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).To(HaveOccurred())
			Expect(cli).To(BeNil())
		})
	})
})
