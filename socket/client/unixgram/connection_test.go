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
	"context"
	"time"

	scksrv "github.com/nabbar/golib/socket/server/unixgram"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UNIX Datagram Client Connection", func() {
	var (
		ctx        context.Context
		cancel     context.CancelFunc
		srv        scksrv.ServerUnixGram
		socketPath string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 10*time.Second)
		socketPath = getTestSocketPath()
		srv = createSimpleTestServer(ctx, socketPath)
	})

	AfterEach(func() {
		if srv != nil && srv.IsRunning() {
			_ = srv.Shutdown(ctx)
		}
		cleanupSocket(socketPath)
		if cancel != nil {
			cancel()
		}
	})

	Describe("Connect", func() {
		Context("with valid server", func() {
			It("should connect successfully", func() {
				cli := createClient(socketPath)
				defer func() {
					_ = cli.Close()
				}()

				err := cli.Connect(ctx)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should mark client as connected", func() {
				cli := createClient(socketPath)
				defer func() {
					_ = cli.Close()
				}()

				Expect(cli.IsConnected()).To(BeFalse())

				connectClient(ctx, cli)
				Expect(cli.IsConnected()).To(BeTrue())
			})

			It("should allow multiple connects (reconnect)", func() {
				cli := createClient(socketPath)
				defer func() {
					_ = cli.Close()
				}()

				err := cli.Connect(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(cli.IsConnected()).To(BeTrue())

				// Reconnect
				err = cli.Connect(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(cli.IsConnected()).To(BeTrue())
			})

			It("should handle concurrent connections", func() {
				cli := createClient(socketPath)
				defer func() {
					_ = cli.Close()
				}()

				done := make(chan bool, 3)
				for i := 0; i < 3; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cli.Connect(ctx)
						done <- true
					}()
				}

				// Wait for all goroutines
				for i := 0; i < 3; i++ {
					Eventually(done, 2*time.Second).Should(Receive())
				}

				Expect(cli.IsConnected()).To(BeTrue())
			})
		})

		Context("with invalid scenarios", func() {
			It("should handle context timeout", func() {
				nonExistentPath := getTestSocketPath()
				cli := createClient(nonExistentPath)
				defer func() {
					_ = cli.Close()
				}()

				timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 50*time.Millisecond)
				defer timeoutCancel()

				err := cli.Connect(timeoutCtx)
				Expect(err).To(HaveOccurred())
			})

			It("should handle connection to non-existent socket", func() {
				nonExistentPath := getTestSocketPath()
				cli := createClient(nonExistentPath)
				defer func() {
					_ = cli.Close()
				}()

				err := cli.Connect(ctx)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("IsConnected", func() {
		It("should return false before connection", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should return true after connection", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)
			Expect(cli.IsConnected()).To(BeTrue())
		})

		It("should return false after close", func() {
			cli := createClient(socketPath)

			connectClient(ctx, cli)
			Expect(cli.IsConnected()).To(BeTrue())

			err := cli.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should be safe to call multiple times", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			for i := 0; i < 10; i++ {
				Expect(cli.IsConnected()).To(BeTrue())
			}
		})
	})

	Describe("Close", func() {
		It("should close connection successfully", func() {
			cli := createClient(socketPath)

			connectClient(ctx, cli)
			Expect(cli.IsConnected()).To(BeTrue())

			err := cli.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should return error when not connected", func() {
			cli := createClient(socketPath)

			err := cli.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should return error on double close", func() {
			cli := createClient(socketPath)

			connectClient(ctx, cli)

			err := cli.Close()
			Expect(err).ToNot(HaveOccurred())

			err = cli.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should be safe to defer close", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)
			Expect(cli.IsConnected()).To(BeTrue())
		})
	})
})
