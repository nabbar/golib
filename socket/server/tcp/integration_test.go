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
	"crypto/tls"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Integration Tests", func() {
	Describe("Full Client-Server Echo Scenario", func() {
		var (
			ctx     context.Context
			cancel  context.CancelFunc
			srv     libsck.Server
			address string
		)

		BeforeEach(func() {
			ctx, cancel = context.WithTimeout(x, 30*time.Second)
			address = getTestAddress()
			srv = createAndRegisterServer(address, echoHandler, nil)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
		})

		AfterEach(func() {
			if srv != nil && srv.IsRunning() {
				_ = srv.Shutdown(ctx)
			}
			if cancel != nil {
				cancel()
			}
		})

		It("should handle complete request-response cycle", func() {
			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)

			// Send multiple messages
			messages := []string{
				"Hello, Server!\n",
				"How are you?\n",
				"Testing 123\n",
			}

			for _, msg := range messages {
				data := []byte(msg)
				n := sendMessage(conn, data)
				Expect(n).To(Equal(len(data)))

				response := receiveMessage(conn, 1024)
				Expect(response).To(Equal(data))
			}

			Expect(srv.OpenConnections()).To(Equal(int64(1)))
		})

		It("should handle graceful disconnect", func() {
			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			msg := []byte("Goodbye\n")
			sendMessage(conn, msg)
			response := receiveMessage(conn, 1024)
			Expect(response).To(Equal(msg))

			_ = conn.Close()

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 10*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Describe("Using Config Package", func() {
		var (
			ctx context.Context
			cnl context.CancelFunc
		)

		BeforeEach(func() {
			ctx, cnl = context.WithTimeout(x, 30*time.Second)
		})

		AfterEach(func() {
			if cnl != nil {
				cnl()
			}
		})

		It("should create server from ServerConfig", func() {
			address := getTestAddress()
			cfg := &sckcfg.ServerConfig{
				Network:   libptc.NetworkTCP,
				Address:   address,
				PermFile:  0,
				GroupPerm: 0,
			}

			srv, err := cfg.New(nil, echoHandler)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)

			_ = srv.Shutdown(ctx)
		})

		It("should create client from ClientConfig and communicate", func() {
			address := getTestAddress()
			srvCfg := &sckcfg.ServerConfig{
				Network: libptc.NetworkTCP,
				Address: address,
			}

			srv, err := srvCfg.New(nil, echoHandler)
			Expect(err).ToNot(HaveOccurred())

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			cliCfg := &sckcfg.ClientConfig{
				Network: libptc.NetworkTCP,
				Address: address,
			}

			cli, err := cliCfg.New()
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())

			err = cli.Connect(ctx)
			defer func() {
				_ = cli.Close()
			}()
			Expect(err).ToNot(HaveOccurred())

			waitForConnections(srv, 1, 2*time.Second)

			msg := []byte("Config Test\n")
			n, err := cli.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(msg)))

			response := make([]byte, len(msg)*2)
			n, err = cli.Read(response)
			Expect(err).ToNot(HaveOccurred())
			Expect(response[:n]).To(Equal(msg))

			_ = srv.Shutdown(ctx)
		})
	})

	Describe("Concurrent Operations", func() {
		var (
			ctx     context.Context
			cancel  context.CancelFunc
			srv     libsck.Server
			address string
		)

		BeforeEach(func() {
			ctx, cancel = context.WithTimeout(x, 60*time.Second)
			address = getTestAddress()
			srv = createAndRegisterServer(address, echoHandler, nil)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
		})

		AfterEach(func() {
			if srv != nil && srv.IsRunning() {
				_ = srv.Shutdown(ctx)
			}
			if cancel != nil {
				cancel()
			}
		})

		It("should handle concurrent clients", func() {
			const numClients = 20
			var wg sync.WaitGroup
			successCount := &atomic.Int32{}

			for i := 0; i < numClients; i++ {
				wg.Add(1)
				go func(clientID int) {
					defer GinkgoRecover()
					defer wg.Done()

					conn, err := net.DialTimeout(libptc.NetworkTCP.Code(), address, 5*time.Second)
					defer func() {
						_ = conn.Close()
					}()

					if err != nil {
						return
					}

					msg := []byte("Client " + string(rune('0'+clientID)) + "\n")
					_, err = conn.Write(msg)
					if err != nil {
						return
					}

					response := make([]byte, len(msg))
					_, err = io.ReadFull(conn, response)
					if err != nil {
						return
					}

					if string(response) == string(msg) {
						successCount.Add(1)
					}
				}(i)
			}

			wg.Wait()

			// Most clients should succeed
			Expect(successCount.Load()).To(BeNumerically(">=", int32(numClients*0.8)))
		})

		It("should handle sequential operations on single connection", func() {
			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)

			// Send and receive multiple messages sequentially
			for i := 0; i < 5; i++ {
				msg := []byte("Test message\n")
				_, err := conn.Write(msg)
				Expect(err).ToNot(HaveOccurred())

				buf := make([]byte, len(msg))
				_, err = io.ReadFull(conn, buf)
				Expect(err).ToNot(HaveOccurred())
				Expect(buf).To(Equal(msg))
			}
		})
	})

	Describe("Long Running Server", func() {
		It("should handle extended operation", func() {
			ctx, cancel := context.WithTimeout(x, 60*time.Second)
			defer cancel()

			address := getTestAddress()
			srv := createAndRegisterServer(address, echoHandler, nil)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			// Run for a while with periodic connections
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			done := time.After(3 * time.Second)
			connCount := 0

		loop:
			for {
				select {
				case <-ticker.C:
					conn := connectClient(address)
					msg := []byte("test\n")
					sendMessage(conn, msg)
					_ = receiveMessage(conn, 1024)
					_ = conn.Close()
					connCount++
				case <-done:
					break loop
				}
			}

			// Should have handled many connections
			Expect(connCount).To(BeNumerically(">", 10))

			// Server should still be running
			Expect(srv.IsRunning()).To(BeTrue())

			// All connections should be closed
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 15*time.Second, 100*time.Millisecond).Should(Equal(int64(0)))

			_ = srv.Shutdown(ctx)
		})
	})

	Describe("TLS Integration", func() {
		var (
			ctx       context.Context
			cancel    context.CancelFunc
			srv       libsck.Server
			address   string
			tlsConfig libtls.TLSConfig
		)

		BeforeEach(func() {
			ctx, cancel = context.WithTimeout(x, 30*time.Second)
			address = getTestAddress()
			srv = createAndRegisterServer(address, echoHandler, nil)
			tlsConfig = createTLSConfig()
			_ = srv.SetTLS(true, tlsConfig)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
		})

		AfterEach(func() {
			if srv != nil && srv.IsRunning() {
				_ = srv.Shutdown(ctx)
			}
			if cancel != nil {
				cancel()
			}
		})

		It("should complete full TLS handshake and communication", func() {
			clientConfig := &tls.Config{
				InsecureSkipVerify: true,
			}

			conn := connectTLSClient(address, clientConfig)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)

			msg := []byte("Secure Message\n")
			n := sendMessage(conn, msg)
			Expect(n).To(Equal(len(msg)))

			response := receiveMessage(conn, 1024)
			Expect(response).To(Equal(msg))
		})

		It("should handle multiple concurrent TLS connections", func() {
			const numConns = 10
			var wg sync.WaitGroup

			for i := 0; i < numConns; i++ {
				wg.Add(1)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()

					clientConfig := &tls.Config{
						InsecureSkipVerify: true,
					}

					conn, err := tls.Dial(libptc.NetworkTCP.Code(), address, clientConfig)
					defer func() {
						_ = conn.Close()
					}()
					if err != nil {
						return
					}

					msg := []byte("TLS Test\n")
					_, _ = conn.Write(msg)

					buf := make([]byte, len(msg))
					_, _ = io.ReadFull(conn, buf)
				}()
			}

			wg.Wait()

			// All connections should eventually close
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 15*time.Second, 100*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Describe("Stress Testing", func() {
		It("should handle rapid connect/disconnect cycles", func() {
			ctx, cancel := context.WithTimeout(x, 60*time.Second)
			defer cancel()

			address := getTestAddress()
			srv := createAndRegisterServer(address, echoHandler, nil)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			const iterations = 100
			for i := 0; i < iterations; i++ {
				conn, err := net.DialTimeout(libptc.NetworkTCP.Code(), address, 2*time.Second)
				if err != nil {
					continue
				}
				_ = conn.Close()
			}

			// Server should still be running
			Expect(srv.IsRunning()).To(BeTrue())

			// All connections should be closed
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 20*time.Second, 100*time.Millisecond).Should(Equal(int64(0)))

			_ = srv.Shutdown(ctx)
		})

		It("should handle large message throughput", func() {
			ctx, cancel := context.WithTimeout(x, 60*time.Second)
			defer cancel()

			address := getTestAddress()
			srv := createAndRegisterServer(address, echoHandler, nil)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)

			// Send many small messages
			const numMessages = 100
			for i := 0; i < numMessages; i++ {
				msg := []byte("Message\n")
				_, err := conn.Write(msg)
				Expect(err).ToNot(HaveOccurred())

				buf := make([]byte, len(msg))
				_, err = io.ReadFull(conn, buf)
				Expect(err).ToNot(HaveOccurred())
			}

			_ = srv.Shutdown(ctx)
		})
	})

	Describe("State Transitions", func() {
		It("should handle start/stop cycle", func() {
			ctx, cancel := context.WithTimeout(x, 30*time.Second)
			defer cancel()

			address := getTestAddress()
			srv := createAndRegisterServer(address, echoHandler, nil)

			// Start
			startServer(ctx, srv)
			waitForServerRunning(srv, 3*time.Second)
			Expect(srv.IsRunning()).To(BeTrue())

			// Connect
			conn := connectClient(address)
			waitForConnections(srv, 1, 3*time.Second)

			// Use
			msg := []byte("Test\n")
			sendMessage(conn, msg)
			_ = receiveMessage(conn, 1024)

			_ = conn.Close()

			// Stop
			_ = srv.Shutdown(ctx)
			waitForServerStopped(srv, 20*time.Second)
			Expect(srv.IsRunning()).To(BeFalse())
		})
	})

	Describe("Error Recovery", func() {
		It("should recover from client errors", func() {
			ctx, cancel := context.WithTimeout(x, 30*time.Second)
			defer cancel()

			errorCount := &atomic.Int32{}
			srv := createAndRegisterServer(getTestAddress(), echoHandler, nil)
			srv.RegisterFuncError(func(errs ...error) {
				errorCount.Add(1)
			})

			address := getTestAddress()
			_ = srv.RegisterServer(address)
			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			// Create and immediately close connections to trigger errors
			for i := 0; i < 5; i++ {
				conn, err := net.DialTimeout(libptc.NetworkTCP.Code(), address, 2*time.Second)
				if err == nil {
					_ = conn.Close()
				}
			}

			time.Sleep(2 * time.Second)

			// Server should still be functional
			Expect(srv.IsRunning()).To(BeTrue())

			// Normal connection should still work
			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			msg := []byte("After errors\n")
			sendMessage(conn, msg)
			response := receiveMessage(conn, 1024)
			Expect(response).To(Equal(msg))

			_ = srv.Shutdown(ctx)
		})
	})
})
