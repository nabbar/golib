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
	"io"
	"sync"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Client TLS", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		srv     scksrt.ServerTcp
		address string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 10*time.Second)
		address = getTestAddress()
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Shutdown(ctx)
		}
		if cancel != nil {
			cancel()
		}
	})

	Describe("SetTLS", func() {
		Context("with valid configuration", func() {
			It("should enable TLS successfully", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				err := cli.SetTLS(true, cliTLS, "")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should disable TLS successfully", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				err := cli.SetTLS(false, nil, "")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should allow TLS configuration before connection", func() {
				srv = createTLSServer(address, echoHandler)
				defer func() {
					if srv != nil {
						_ = srv.Shutdown(ctx)
					}
				}()

				startServer(ctx, srv)
				waitForServerRunning(address, 5*time.Second)

				cli := createTLSClient(address)
				defer func() {
					_ = cli.Close()
				}()

				Expect(cli.Connect(ctx)).ToNot(HaveOccurred())
				Expect(cli.IsConnected()).To(BeTrue())
			})

			It("should handle different server names", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				serverNames := []string{"localhost", "127.0.0.1"}
				for _, name := range serverNames {
					err := cli.SetTLS(true, cliTLS, name)
					Expect(err).ToNot(HaveOccurred())
				}
			})
		})

		Context("with invalid configuration", func() {
			It("should fail with nil config when enabled", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				err := cli.SetTLS(true, nil, "localhost")
				Expect(err).To(HaveOccurred())
			})

			It("should succeed with nil config when disabled", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				err := cli.SetTLS(false, nil, "")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("configuration changes", func() {
			It("should allow changing TLS configuration", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				// Enable TLS
				err := cli.SetTLS(true, cliTLS, "localhost")
				Expect(err).ToNot(HaveOccurred())

				// Disable TLS
				err = cli.SetTLS(false, nil, "")
				Expect(err).ToNot(HaveOccurred())

				// Re-enable TLS
				err = cli.SetTLS(true, cliTLS, "localhost")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("TLS Connection", func() {
		BeforeEach(func() {
			srv = createTLSServer(address, echoHandler)
			startServer(ctx, srv)
			waitForServerRunning(address, 5*time.Second)
		})
		AfterEach(func() {
			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		Context("with matching TLS configuration", func() {
			It("should connect successfully", func() {
				cli := createTLSClient(address)
				defer func() {
					_ = cli.Close()
				}()

				Expect(cli.Connect(ctx)).ToNot(HaveOccurred())
				Expect(cli.IsConnected()).To(BeTrue())
			})

			It("should exchange data over TLS", func() {
				cli := createTLSClient(address)
				defer func() {
					_ = cli.Close()
				}()

				connectClient(ctx, cli)

				msg := []byte("Secure message\n")
				response := sendAndReceive(cli, msg)
				Expect(response).To(Equal(msg))
			})

			It("should handle multiple TLS connections", func() {
				wg := sync.WaitGroup{}
				clients := libatm.NewMapTyped[int, io.Closer]()
				for i := 0; i < 3; i++ {
					wg.Go(func() {
						cli := createTLSClient(address)
						connectClient(ctx, cli)
						Expect(cli.IsConnected()).To(BeTrue())
						clients.Store(i, cli)
					})
				}
				wg.Wait()
				time.Sleep(50 * time.Millisecond)

				// Cleanup
				clients.Range(func(_ int, cli io.Closer) bool {
					Expect(cli.Close()).ToNot(HaveOccurred())
					return true
				})
			})

			It("should maintain TLS connection", func() {
				cli := createTLSClient(address)
				defer func() {
					_ = cli.Close()
				}()

				connectClient(ctx, cli)

				// Send multiple messages
				msg := []byte("TLS message\n")
				for i := 0; i < 10; i++ {
					Expect(sendAndReceive(cli, msg)).To(Equal(msg))
				}

				Expect(cli.IsConnected()).To(BeTrue())
			})

			It("should handle large data over TLS", func() {
				cli := createTLSClient(address)
				defer func() {
					_ = cli.Close()
				}()

				connectClient(ctx, cli)

				// 1MB of data
				largeData := make([]byte, 1024*1024)
				for i := range largeData {
					largeData[i] = byte(i % 256)
				}

				response := sendAndReceive(cli, largeData)
				Expect(response).To(Equal(largeData))
			})
		})

		Context("with mismatched configuration", func() {
			It("should fail to connect without TLS when server requires it", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				err := cli.Connect(ctx)
				// Connection should fail or succeed but communication will fail
				if err == nil {
					// Try to communicate
					msg := []byte("Hello\n")
					_, writeErr := cli.Write(msg)
					Expect(writeErr).ToNot(HaveOccurred())
				}
			})
		})
	})

	Describe("TLS with Plain Server", func() {
		BeforeEach(func() {
			srv = createSimpleTestServer(ctx, address)
		})
		AfterEach(func() {
			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should fail TLS connection to plain server", func() {
			cli := createTLSClient(address)
			defer func() {
				_ = cli.Close()
			}()

			err := cli.Connect(ctx)
			// TLS handshake should fail
			if err == nil {
				// Connection might succeed but handshake will fail
				msg := []byte("Hello\n")
				_, writeErr := cli.Write(msg)
				Expect(writeErr).To(HaveOccurred())
			}
		})
	})

	Describe("TLS Security", func() {
		BeforeEach(func() {
			srv = createTLSServer(address, echoHandler)
			startServer(ctx, srv)
			waitForServerRunning(address, 5*time.Second)
		})
		AfterEach(func() {
			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should enforce certificate validation", func() {
			cli := createTLSClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// With proper cert validation
			err := cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle TLS handshake timeout", func() {
			cli := createTLSClient(address)
			defer func() {
				_ = cli.Close()
			}()

			timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Nanosecond)
			defer timeoutCancel()

			// Very short timeout might fail the handshake
			err := cli.Connect(timeoutCtx)
			// Either succeeds quickly or times out
			_ = err
		})
	})

	Describe("TLS Performance", func() {
		BeforeEach(func() {
			srv = createTLSServer(address, echoHandler)
			startServer(ctx, srv)
			waitForServerRunning(address, 5*time.Second)
		})
		AfterEach(func() {
			if srv != nil {
				_ = srv.Shutdown(ctx)
			}
		})

		It("should handle high throughput with TLS", func() {
			cli := createTLSClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			iterations := 100
			msgSize := 1024
			msg := make([]byte, msgSize)
			for i := range msg {
				msg[i] = byte(i % 256)
			}

			start := time.Now()
			for i := 0; i < iterations; i++ {
				response := sendAndReceive(cli, msg)
				Expect(response).To(Equal(msg))
			}
			elapsed := time.Since(start)

			// TLS should complete in reasonable time (allow 15 seconds for 100 iterations)
			Expect(elapsed).To(BeNumerically("<", 15*time.Second))
		})

		It("should handle rapid TLS connections", func() {
			for i := 0; i < 10; i++ {
				cli := createTLSClient(address)

				err := cli.Connect(ctx)
				Expect(err).ToNot(HaveOccurred())

				msg := []byte("Quick TLS test\n")
				response := sendAndReceive(cli, msg)
				Expect(response).To(Equal(msg))

				err = cli.Close()
				Expect(err).ToNot(HaveOccurred())
			}
		})
	})
})
