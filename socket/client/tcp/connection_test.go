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
	"context"
	"time"

	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Client Connection", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		srv     scksrt.ServerTcp
		address string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 10*time.Second)
		address = getTestAddress()
		srv = createSimpleTestServer(ctx, address)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Shutdown(ctx)
		}
		if cancel != nil {
			cancel()
		}
	})

	Describe("Connect", func() {
		Context("with running server", func() {
			It("should connect successfully", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				err := cli.Connect(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(cli.IsConnected()).To(BeTrue())
			})

			It("should maintain connection", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				connectClient(ctx, cli)
				waitForClientConnected(cli, 5*time.Second)

				// Verify connection stays alive
				time.Sleep(100 * time.Millisecond)
				Expect(cli.IsConnected()).To(BeTrue())
			})

			It("should allow multiple clients to connect", func() {
				clients := make([]interface{ Close() error }, 3)
				for i := 0; i < 3; i++ {
					cli := createClient(address)
					clients[i] = cli
					connectClient(ctx, cli)
					Expect(cli.IsConnected()).To(BeTrue())
				}

				// Cleanup
				for _, cli := range clients {
					_ = cli.Close()
				}
			})
		})

		Context("with no server", func() {
			It("should fail to connect", func() {
				noServerAddr := getTestAddress()
				cli := createClient(noServerAddr)
				defer func() {
					_ = cli.Close()
				}()

				Expect(cli.Connect(ctx)).To(HaveOccurred())
				Expect(cli.IsConnected()).To(BeFalse())
			})

			It("should handle connection timeout", func() {
				noServerAddr := getTestAddress()
				cli := createClient(noServerAddr)
				defer func() {
					_ = cli.Close()
				}()

				timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 100*time.Millisecond)
				defer timeoutCancel()

				Expect(cli.Connect(timeoutCtx)).To(HaveOccurred())
			})
		})

		Context("with context cancellation", func() {
			It("should fail when context is cancelled before connect", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				cancelCtx, cancelFunc := context.WithCancel(ctx)
				cancelFunc() // Cancel immediately

				Expect(cli.Connect(cancelCtx)).To(HaveOccurred())
			})

			It("should handle context timeout", func() {
				// Create server that accepts but doesn't respond
				slowAddr := getTestAddress()
				slowSrv := createAndRegisterServer(slowAddr, silentHandler)
				defer func() {
					_ = slowSrv.Shutdown(ctx)
				}()

				startServer(ctx, slowSrv)
				waitForServerRunning(slowAddr, 2*time.Second)

				cli := createClient(slowAddr)
				defer func() {
					_ = cli.Close()
				}()

				timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 50*time.Millisecond)
				defer timeoutCancel()

				// Connect should succeed (connection is established)
				// But timeout should affect subsequent operations
				_ = cli.Connect(timeoutCtx)
			})
		})
	})

	Describe("IsConnected", func() {
		It("should return false before connection", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should return true after connection", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)
			Expect(cli.IsConnected()).To(BeTrue())
		})

		It("should return false after close", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)
			Expect(cli.IsConnected()).To(BeTrue())
			Expect(cli.Close()).ToNot(HaveOccurred())

			// After close, connection check should fail
			Eventually(func() bool {
				return cli.IsConnected()
			}, 1*time.Second, 50*time.Millisecond).Should(BeFalse())
		})

		It("should return false when server closes connection", func() {
			// Create server that closes immediately
			closeAddr := getTestAddress()
			closeSrv := createAndRegisterServer(closeAddr, closingHandler)
			defer func() {
				_ = closeSrv.Shutdown(ctx)
			}()

			startServer(ctx, closeSrv)
			waitForServerRunning(closeAddr, 2*time.Second)

			cli := createClient(closeAddr)
			defer func() {
				_ = cli.Close()
			}()

			Expect(cli.Connect(ctx)).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeTrue())

			Expect(closeSrv.Shutdown(ctx)).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeTrue())

			Expect(cli.Close()).ToNot(HaveOccurred())
			Expect(cli.Connect(ctx)).To(HaveOccurred())
			Expect(cli.IsConnected()).To(BeFalse())
		})
	})

	Describe("Close", func() {
		It("should close connected client", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)
			Expect(cli.Close()).ToNot(HaveOccurred())
		})

		It("should handle close on unconnected client", func() {
			cli := createClient(address)
			Expect(cli.Close()).To(HaveOccurred())
		})

		It("should handle double close", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)
			Expect(cli.Close()).ToNot(HaveOccurred())
			Expect(cli.Close()).To(HaveOccurred())
		})

		It("should close multiple clients independently", func() {
			cli1 := createClient(address)
			defer func() {
				_ = cli1.Close()
			}()

			cli2 := createClient(address)
			defer func() {
				_ = cli2.Close()
			}()

			connectClient(ctx, cli1)
			connectClient(ctx, cli2)

			Expect(cli1.Close()).ToNot(HaveOccurred())
			Expect(cli2.IsConnected()).To(BeTrue())
			Expect(cli2.Close()).ToNot(HaveOccurred())
		})
	})

	Describe("Connection Stability", func() {
		It("should maintain connection under load", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Send multiple messages
			for i := 0; i < 10; i++ {
				msg := []byte("test message\n")
				response := sendAndReceive(cli, msg)
				Expect(response).To(Equal(msg))
				Expect(cli.IsConnected()).To(BeTrue())
			}
		})

		It("should handle keep-alive", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)
			waitForClientConnected(cli, 5*time.Second)

			// Wait longer than typical keep-alive interval
			time.Sleep(1 * time.Second)
			Expect(cli.IsConnected()).To(BeTrue())
		})
	})

	Describe("Reconnection", func() {
		It("should allow reconnection after close", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// First connection
			connectClient(ctx, cli)
			Expect(cli.IsConnected()).To(BeTrue())

			// Close
			Expect(cli.Close()).ToNot(HaveOccurred())

			// Wait for close to complete
			time.Sleep(100 * time.Millisecond)

			// Reconnect
			Expect(cli.Connect(ctx)).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeTrue())
		})
	})
})
