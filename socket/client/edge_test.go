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

var _ = Describe("Edge Cases", func() {
	Context("Invalid Configurations", func() {
		It("should handle empty address", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).To(HaveOccurred())
			Expect(cli).To(BeNil())
		})

		It("should handle malformed address", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "invalid::address::format",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).To(HaveOccurred())
			Expect(cli).To(BeNil())
		})

		It("should handle zero protocol value", func() {
			cfg := sckcfg.Client{
				Network: 0,
				Address: "localhost:8080",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).To(HaveOccurred())
			Expect(cli).To(BeNil())
		})
	})

	Context("Boundary Values", func() {
		It("should handle port 0 (system-assigned)", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:0",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
			defer cli.Close()
		})

		It("should handle maximum valid port", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:65535",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
			defer cli.Close()
		})
	})

	Context("Special Addresses", func() {
		It("should handle localhost", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
			defer cli.Close()
		})

		It("should handle 127.0.0.1", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "127.0.0.1:8080",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
			defer cli.Close()
		})

		It("should handle IPv6 loopback", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "[::1]:8080",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
			defer cli.Close()
		})
	})

	Context("Multiple Close Calls", func() {
		It("should handle multiple close calls gracefully", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			// First close
			err = cli.Close()
			// Close may return error if not connected
			_ = err

			// Second close - should not panic
			err = cli.Close()
			// Error or success both acceptable
			_ = err
		})
	})
})
