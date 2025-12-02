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
	"os"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libptc "github.com/nabbar/golib/network/protocol"
	sckclt "github.com/nabbar/golib/socket/client"
	sckcfg "github.com/nabbar/golib/socket/config"
)

var _ = Describe("Client Factory Creation", func() {
	Context("TCP Client Creation", func() {
		It("should create TCP client successfully", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			if cli != nil {
				_ = cli.Close()
			}
		})

		It("should create TCP4 client successfully", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP4,
				Address: "127.0.0.1:8081",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			if cli != nil {
				_ = cli.Close()
			}
		})

		It("should create TCP6 client successfully", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP6,
				Address: "[::1]:8082",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			if cli != nil {
				_ = cli.Close()
			}
		})
	})

	Context("UDP Client Creation", func() {
		It("should create UDP client successfully", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkUDP,
				Address: "localhost:9000",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			if cli != nil {
				_ = cli.Close()
			}
		})

		It("should create UDP4 client successfully", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkUDP4,
				Address: "127.0.0.1:9001",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			if cli != nil {
				_ = cli.Close()
			}
		})

		It("should create UDP6 client successfully", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkUDP6,
				Address: "[::1]:9002",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			if cli != nil {
				_ = cli.Close()
			}
		})
	})

	Context("Unix Socket Client Creation", func() {
		// Unix sockets are only available on Linux and Darwin
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			It("should create Unix client successfully", func() {
				socketPath := getTestUnixPath()
				defer os.Remove(socketPath)

				cfg := sckcfg.Client{
					Network: libptc.NetworkUnix,
					Address: socketPath,
				}

				cli, err := sckclt.New(cfg, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(cli).ToNot(BeNil())

				if cli != nil {
					_ = cli.Close()
				}
			})

			It("should create UnixGram client successfully", func() {
				socketPath := getTestUnixGramPath()
				defer os.Remove(socketPath)

				cfg := sckcfg.Client{
					Network: libptc.NetworkUnixGram,
					Address: socketPath,
				}

				cli, err := sckclt.New(cfg, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(cli).ToNot(BeNil())

				if cli != nil {
					_ = cli.Close()
				}
			})
		} else {
			It("should return error for Unix client on unsupported platform", func() {
				socketPath := getTestUnixPath()

				cfg := sckcfg.Client{
					Network: libptc.NetworkUnix,
					Address: socketPath,
				}

				cli, err := sckclt.New(cfg, nil)
				Expect(err).To(Equal(sckcfg.ErrInvalidProtocol))
				Expect(cli).To(BeNil())
			})

			It("should return error for UnixGram client on unsupported platform", func() {
				socketPath := getTestUnixGramPath()

				cfg := sckcfg.Client{
					Network: libptc.NetworkUnixGram,
					Address: socketPath,
				}

				cli, err := sckclt.New(cfg, nil)
				Expect(err).To(Equal(sckcfg.ErrInvalidProtocol))
				Expect(cli).To(BeNil())
			})
		}
	})

	Context("Error Handling", func() {
		It("should return error for invalid protocol", func() {
			cfg := sckcfg.Client{
				Network: 255, // Invalid protocol
				Address: "localhost:8080",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).To(HaveOccurred())
			Expect(cli).To(BeNil())
		})

		It("should return error for empty address", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "", // Empty address
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).To(HaveOccurred())
			Expect(cli).To(BeNil())
		})

		It("should return error for unrecognized protocol value", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkProtocol(255), // Unrecognized protocol
				Address: "localhost:8080",
			}

			cli, err := sckclt.New(cfg, nil)
			Expect(err).To(HaveOccurred())
			Expect(cli).To(BeNil())
		})
	})

	Context("TLS Configuration", func() {
		It("should create TCP client with TLS enabled", func() {
			cfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8443",
			}
			cfg.TLS.Enabled = true
			cfg.TLS.ServerName = "localhost"

			cli, err := sckclt.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			if cli != nil {
				_ = cli.Close()
			}
		})
	})

	Context("Concurrent Creation", func() {
		It("should allow concurrent client creation", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func(idx int) {
					defer GinkgoRecover()

					cfg := sckcfg.Client{
						Network: libptc.NetworkTCP,
						Address: "localhost:8080",
					}

					cli, err := sckclt.New(cfg, nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(cli).ToNot(BeNil())

					if cli != nil {
						_ = cli.Close()
					}

					done <- true
				}(i)
			}

			// Wait for all goroutines to complete
			for i := 0; i < 10; i++ {
				Eventually(done, 5*time.Second).Should(Receive())
			}
		})
	})

	Context("Multiple Clients", func() {
		It("should allow creating multiple clients with different protocols", func() {
			// TCP client
			tcpCfg := sckcfg.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8100",
			}
			tcpCli, err := sckclt.New(tcpCfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(tcpCli).ToNot(BeNil())

			// UDP client
			udpCfg := sckcfg.Client{
				Network: libptc.NetworkUDP,
				Address: "localhost:9100",
			}
			udpCli, err := sckclt.New(udpCfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(udpCli).ToNot(BeNil())

			// Cleanup
			if tcpCli != nil {
				_ = tcpCli.Close()
			}
			if udpCli != nil {
				_ = udpCli.Close()
			}
		})
	})
})
