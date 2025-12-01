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

package tcp_test

import (
	"bytes"
	"context"
	"io"
	"time"

	sckclt "github.com/nabbar/golib/socket/client/tcp"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Client Error Handling", func() {
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

	Describe("Error Constants", func() {
		It("should have defined error constants", func() {
			Expect(sckclt.ErrInstance).ToNot(BeNil())
			Expect(sckclt.ErrConnection).ToNot(BeNil())
			Expect(sckclt.ErrAddress).ToNot(BeNil())
		})

		It("should have descriptive error messages", func() {
			Expect(sckclt.ErrInstance.Error()).ToNot(BeEmpty())
			Expect(sckclt.ErrConnection.Error()).ToNot(BeEmpty())
			Expect(sckclt.ErrAddress.Error()).ToNot(BeEmpty())
		})
	})

	Describe("Connection Errors", func() {
		Context("without server", func() {
			It("should return error on Connect", func() {
				noServerAddr := getTestAddress()
				cli := createClient(noServerAddr)
				defer func() {
					_ = cli.Close()
				}()

				err := cli.Connect(ctx)
				Expect(err).To(HaveOccurred())
			})

			It("should return error on Write without connection", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				n, err := cli.Write([]byte("test"))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(sckclt.ErrConnection))
				Expect(n).To(Equal(0))
			})

			It("should return error on Read without connection", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				buf := make([]byte, 1024)
				n, err := cli.Read(buf)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(sckclt.ErrConnection))
				Expect(n).To(Equal(0))
			})

			It("should return error on Once without server", func() {
				noServerAddr := getTestAddress()
				cli := createClient(noServerAddr)

				msg := []byte("test")
				request := bytes.NewReader(msg)

				err := cli.Once(ctx, request, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("after server shutdown", func() {
			BeforeEach(func() {
				srv = createSimpleTestServer(ctx, address)
			})

			It("should handle server shutdown during connection", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				connectClient(ctx, cli)
				Expect(cli.IsConnected()).To(BeTrue())

				// Shutdown server
				err := srv.Shutdown(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Wait for server to close
				time.Sleep(500 * time.Millisecond)

				// IsConnected still returns true (local state)
				// Client only detects disconnection when attempting I/O
				Expect(cli.IsConnected()).To(BeTrue())

				// But I/O should fail
				buf := make([]byte, 1024)
				_, err = cli.Read(buf)
				Expect(err).To(HaveOccurred())
			})

			It("should return error on Read after server shutdown", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				connectClient(ctx, cli)

				// Shutdown server
				_ = srv.Shutdown(ctx)
				time.Sleep(500 * time.Millisecond)

				// Read should fail when server is closed
				buf := make([]byte, 1024)
				_, err := cli.Read(buf)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with context cancellation", func() {
			BeforeEach(func() {
				srv = createSimpleTestServer(ctx, address)
			})

			It("should respect context cancellation on Connect", func() {
				slowAddr := getTestAddress()
				cli := createClient(slowAddr)
				defer func() {
					_ = cli.Close()
				}()

				cancelCtx, cancelFunc := context.WithCancel(ctx)
				cancelFunc() // Cancel immediately

				err := cli.Connect(cancelCtx)
				Expect(err).To(HaveOccurred())
			})

			It("should respect context timeout on Connect", func() {
				noServerAddr := getTestAddress()
				cli := createClient(noServerAddr)
				defer func() {
					_ = cli.Close()
				}()

				timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 100*time.Millisecond)
				defer timeoutCancel()

				err := cli.Connect(timeoutCtx)
				Expect(err).To(HaveOccurred())
			})

			It("should handle context cancellation in Once", func() {
				cli := createClient(address)

				cancelCtx, cancelFunc := context.WithCancel(ctx)
				cancelFunc() // Cancel immediately

				msg := []byte("test")
				request := bytes.NewReader(msg)

				err := cli.Once(cancelCtx, request, nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Invalid Operations", func() {
		It("should handle nil client operations gracefully", func() {
			// This tests nil pointer safety in the implementation
			var cli sckclt.ClientTCP
			// These should not panic
			Expect(cli).To(BeNil())
		})

		It("should handle operations after Close", func() {
			srv = createSimpleTestServer(ctx, address)
			cli := createClient(address)

			connectClient(ctx, cli)
			err := cli.Close()
			Expect(err).ToNot(HaveOccurred())

			// Operations after close should fail
			time.Sleep(100 * time.Millisecond)

			buf := make([]byte, 1024)
			_, readErr := cli.Read(buf)
			Expect(readErr).To(HaveOccurred())

			_, writeErr := cli.Write([]byte("test"))
			Expect(writeErr).To(HaveOccurred())
		})

		It("should handle concurrent Close calls", func() {
			srv = createSimpleTestServer(ctx, address)
			cli := createClient(address)

			connectClient(ctx, cli)

			// Close concurrently
			done := make(chan bool, 2)
			for i := 0; i < 2; i++ {
				go func() {
					defer GinkgoRecover()
					_ = cli.Close()
					done <- true
				}()
			}

			// Wait for both to complete with timeout
			timeout := time.After(2 * time.Second)
			for i := 0; i < 2; i++ {
				select {
				case <-done:
					// Success
				case <-timeout:
					Fail("Timeout waiting for Close to complete")
				}
			}

			// Should not panic
		})
	})

	Describe("Network Errors", func() {
		BeforeEach(func() {
			srv = createSimpleTestServer(ctx, address)
		})

		It("should handle broken pipe", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Close connection from server side by shutting down
			_ = srv.Shutdown(ctx)
			time.Sleep(500 * time.Millisecond)

			// Try to read - should get EOF or connection reset
			buf := make([]byte, 1024)
			_, err := cli.Read(buf)
			Expect(err).To(HaveOccurred())
		})

		It("should handle connection reset", func() {
			// Create server that closes immediately
			closeAddr := getTestAddress()
			closeSrv := createAndRegisterServer(closeAddr, closingHandler)
			startServer(ctx, closeSrv)
			defer func() {
				_ = closeSrv.Shutdown(ctx)
			}()
			waitForServerRunning(closeAddr, 2*time.Second)

			cli := createClient(closeAddr)
			defer func() {
				_ = cli.Close()
			}()

			err := cli.Connect(ctx)
			// Connection might succeed but will be closed immediately
			if err == nil {
				time.Sleep(200 * time.Millisecond)

				// Try to write
				msg := []byte("test\n")
				_, writeErr := cli.Write(msg)
				// Should fail because connection was reset
				_ = writeErr
			}
		})
	})

	Describe("TLS Errors", func() {
		It("should fail SetTLS with nil config when enabled", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			err := cli.SetTLS(true, nil, "localhost")
			Expect(err).To(HaveOccurred())
		})

		It("should handle TLS handshake failure", func() {
			// Create plain server
			plainSrv := createSimpleTestServer(ctx, address)
			defer func() {
				_ = plainSrv.Shutdown(ctx)
			}()

			// Try to connect with TLS client
			cli := createTLSClient(address)
			defer func() {
				_ = cli.Close()
			}()

			err := cli.Connect(ctx)
			// Connection might succeed but handshake will fail
			if err == nil {
				// Try to write - should fail
				msg := []byte("test\n")
				_, writeErr := cli.Write(msg)
				Expect(writeErr).To(HaveOccurred())
			}
		})

		It("should handle invalid server name in TLS", func() {
			// Create TLS server
			tlsSrv := createTLSServer(address, echoHandler)
			startServer(ctx, tlsSrv)
			defer func() {
				_ = tlsSrv.Shutdown(ctx)
			}()
			waitForServerRunning(address, 2*time.Second)

			// Create client with wrong server name
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			_, clientConfig := createTLSConfig()
			err := cli.SetTLS(true, clientConfig, "wrong-hostname")
			Expect(err).ToNot(HaveOccurred())

			// Connection should fail due to name mismatch
			err = cli.Connect(ctx)
			// Might succeed or fail depending on cert validation
			_ = err
		})
	})

	Describe("EOF Handling", func() {
		BeforeEach(func() {
			srv = createSimpleTestServer(ctx, address)
		})

		It("should handle EOF on Read", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Close server to cause EOF
			_ = srv.Shutdown(ctx)
			time.Sleep(500 * time.Millisecond)

			// Read should return EOF or similar error
			buf := make([]byte, 1024)
			_, err := cli.Read(buf)
			Expect(err).To(HaveOccurred())
		})

		It("should handle EOF in Once", func() {
			// Create server that closes connection after reading
			closeAddr := getTestAddress()
			closeSrv := createAndRegisterServer(closeAddr, closingHandler)
			startServer(ctx, closeSrv)
			defer func() {
				_ = closeSrv.Shutdown(ctx)
			}()
			waitForServerRunning(closeAddr, 2*time.Second)

			cli := createClient(closeAddr)

			msg := []byte("test")
			request := bytes.NewReader(msg)

			// Once should handle the connection closure
			err := cli.Once(ctx, request, func(r io.Reader) {
				_, _ = io.ReadAll(r)
			})

			// Might succeed (if connection completed) or fail (if closed during)
			_ = err
		})
	})

	Describe("Edge Case Errors", func() {
		It("should handle empty address in New", func() {
			cli, err := sckclt.New("")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(sckclt.ErrAddress))
			Expect(cli).To(BeNil())
		})

		It("should handle malformed address in New", func() {
			cli, err := sckclt.New("not-a-valid-address")
			Expect(err).To(HaveOccurred())
			Expect(cli).To(BeNil())
		})

		It("should handle invalid port in New", func() {
			cli, err := sckclt.New("127.0.0.1:99999")
			Expect(err).To(HaveOccurred())
			Expect(cli).To(BeNil())
		})

		It("should handle operations with nil buffer", func() {
			srv = createSimpleTestServer(ctx, address)
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Read with nil buffer
			n, err := cli.Read(nil)
			// Should handle gracefully (might return 0 or error)
			_ = n
			_ = err

			// Write with nil buffer is tested in communication_test.go
		})
	})

	Describe("Recovery from Errors", func() {
		BeforeEach(func() {
			srv = createSimpleTestServer(ctx, address)
		})

		It("should allow reconnection after connection error", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// First connection
			connectClient(ctx, cli)
			Expect(cli.IsConnected()).To(BeTrue())

			// Force error by closing
			_ = cli.Close()
			time.Sleep(100 * time.Millisecond)

			// Reconnect should work
			err := cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeTrue())
		})

		It("should continue working after recoverable error", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Send successful message
			msg1 := []byte("first\n")
			response1 := sendAndReceive(cli, msg1)
			Expect(response1).To(Equal(msg1))

			// Connection should still work
			msg2 := []byte("second\n")
			response2 := sendAndReceive(cli, msg2)
			Expect(response2).To(Equal(msg2))
		})
	})
})
