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

package server_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server"
)

var _ = Describe("Server Factory Creation", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
		time.Sleep(50 * time.Millisecond) // Allow cleanup
	})

	Context("TCP Server Creation", func() {
		It("should create TCP server successfully", func() {
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

		It("should create TCP4 server successfully", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP4,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should create TCP6 server successfully", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkTCP6,
				Address: getTestTCPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})
	})

	Context("UDP Server Creation", func() {
		It("should create UDP server successfully", func() {
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

		It("should create UDP4 server successfully", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkUDP4,
				Address: getTestUDPAddress(),
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should create UDP6 server successfully", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkUDP6,
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

	Context("Unix Socket Server Creation", func() {
		// Unix sockets are only available on Linux and Darwin
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			It("should create Unix server successfully", func() {
				tmpDir := os.TempDir()
				socketPath := filepath.Join(tmpDir, fmt.Sprintf("test-unix-%d.sock", time.Now().UnixNano()))
				defer os.Remove(socketPath)

				cfg := sckcfg.Server{
					Network:   libptc.NetworkUnix,
					Address:   socketPath,
					PermFile:  libprm.Perm(0600),
					GroupPerm: -1,
				}

				srv, err := scksrv.New(nil, basicHandler(), cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(srv).ToNot(BeNil())

				if srv != nil {
					_ = srv.Shutdown(ctx)
				}
			})

			It("should create UnixGram server successfully", func() {
				tmpDir := os.TempDir()
				socketPath := filepath.Join(tmpDir, fmt.Sprintf("test-unixgram-%d.sock", time.Now().UnixNano()))
				defer os.Remove(socketPath)

				cfg := sckcfg.Server{
					Network:   libptc.NetworkUnixGram,
					Address:   socketPath,
					PermFile:  libprm.Perm(0600),
					GroupPerm: -1,
				}

				srv, err := scksrv.New(nil, basicHandler(), cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(srv).ToNot(BeNil())

				if srv != nil {
					_ = srv.Shutdown(ctx)
				}
			})
		} else {
			It("should return error for Unix server on unsupported platform", func() {
				tmpDir := os.TempDir()
				socketPath := filepath.Join(tmpDir, fmt.Sprintf("test-unix-%d.sock", time.Now().UnixNano()))

				cfg := sckcfg.Server{
					Network:   libptc.NetworkUnix,
					Address:   socketPath,
					PermFile:  libprm.Perm(0600),
					GroupPerm: -1,
				}

				srv, err := scksrv.New(nil, basicHandler(), cfg)
				Expect(err).To(Equal(sckcfg.ErrInvalidProtocol))
				Expect(srv).To(BeNil())
			})

			It("should return error for UnixGram server on unsupported platform", func() {
				tmpDir := os.TempDir()
				socketPath := filepath.Join(tmpDir, fmt.Sprintf("test-unixgram-%d.sock", time.Now().UnixNano()))

				cfg := sckcfg.Server{
					Network:   libptc.NetworkUnixGram,
					Address:   socketPath,
					PermFile:  libprm.Perm(0600),
					GroupPerm: -1,
				}

				srv, err := scksrv.New(nil, basicHandler(), cfg)
				Expect(err).To(Equal(sckcfg.ErrInvalidProtocol))
				Expect(srv).To(BeNil())
			})
		}
	})

	Context("Error Handling", func() {
		It("should return error for invalid protocol", func() {
			cfg := sckcfg.Server{
				Network: 255, // Invalid protocol
				Address: ":8080",
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).To(Equal(sckcfg.ErrInvalidProtocol))
			Expect(srv).To(BeNil())
		})

		It("should return error for unrecognized protocol value", func() {
			cfg := sckcfg.Server{
				Network: libptc.NetworkProtocol(255), // Unrecognized protocol
				Address: ":8080",
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).To(Equal(sckcfg.ErrInvalidProtocol))
			Expect(srv).To(BeNil())
		})
	})

	Context("Configuration Options", func() {
		It("should create server with idle timeout", func() {
			cfg := sckcfg.Server{
				Network:        libptc.NetworkTCP,
				Address:        getTestTCPAddress(),
				ConIdleTimeout: 5 * time.Minute,
			}

			srv, err := scksrv.New(nil, basicHandler(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should create server with custom handler", func() {
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

	Context("Concurrent Creation", func() {
		It("should allow concurrent server creation", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()

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

					done <- true
				}()
			}

			// Wait for all goroutines to complete
			for i := 0; i < 10; i++ {
				Eventually(done, 5*time.Second).Should(Receive())
			}
		})
	})

	Context("Multiple Servers", func() {
		It("should allow creating multiple servers with different protocols", func() {
			// TCP server
			tcpCfg := sckcfg.Server{
				Network: libptc.NetworkTCP,
				Address: getTestTCPAddress(),
			}
			tcpSrv, err := scksrv.New(nil, basicHandler(), tcpCfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(tcpSrv).ToNot(BeNil())

			// UDP server
			udpCfg := sckcfg.Server{
				Network: libptc.NetworkUDP,
				Address: getTestUDPAddress(),
			}
			udpSrv, err := scksrv.New(nil, basicHandler(), udpCfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(udpSrv).ToNot(BeNil())

			// Cleanup
			if tcpSrv != nil {
				_ = tcpSrv.Shutdown(ctx)
			}
			if udpSrv != nil {
				_ = udpSrv.Shutdown(ctx)
			}
		})
	})
})
