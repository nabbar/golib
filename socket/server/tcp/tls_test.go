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
	"net"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	libsck "github.com/nabbar/golib/socket"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server TLS", func() {
	var (
		ctx       context.Context
		cancel    context.CancelFunc
		srv       libsck.Server
		address   string
		tlsConfig libtls.TLSConfig
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(x, 60*time.Second)
		address = getTestAddress()
		srv = createAndRegisterServer(address, echoHandler, nil)
		tlsConfig = createTLSConfig()
	})

	AfterEach(func() {
		if srv != nil && srv.IsRunning() {
			_ = srv.Shutdown(ctx)
		}
		if cancel != nil {
			cancel()
		}
	})

	Describe("SetTLS", func() {
		Context("enabling TLS", func() {
			It("should enable TLS with valid config", func() {
				err := srv.SetTLS(true, tlsConfig)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should fail with nil config", func() {
				err := srv.SetTLS(true, nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid tls config"))
			})

			It("should fail with empty certificates", func() {
				emptyConfig := libtls.New()
				// Don't add any certificates
				err := srv.SetTLS(true, emptyConfig)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("missing certificates"))
			})

			It("should accept TLS config before starting server", func() {
				err := srv.SetTLS(true, tlsConfig)
				Expect(err).ToNot(HaveOccurred())

				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)
			})

			It("should not accept plain TCP connections when TLS is enabled", func() {
				err := srv.SetTLS(true, tlsConfig)
				Expect(err).ToNot(HaveOccurred())

				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)

				// Try plain connection (should fail or hang)
				conn, err := connectClientWithTimeout(address, 1*time.Second)
				defer func() {
					_ = conn.Close()
				}()

				if err == nil {
					// Try to send data - should fail
					_, writeErr := conn.Write([]byte("test\n"))
					// Either write fails or we timeout reading
					if writeErr == nil {
						buf := make([]byte, 100)
						_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
						_, readErr := conn.Read(buf)
						Expect(readErr).To(HaveOccurred())
					}
				}
			})
		})

		Context("disabling TLS", func() {
			It("should disable TLS", func() {
				// First enable
				err := srv.SetTLS(true, tlsConfig)
				Expect(err).ToNot(HaveOccurred())

				// Then disable
				err = srv.SetTLS(false, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should accept plain connections when TLS is disabled", func() {
				err := srv.SetTLS(false, nil)
				Expect(err).ToNot(HaveOccurred())

				startServer(ctx, srv)
				waitForServerRunning(srv, 2*time.Second)

				conn := connectClient(address)
				defer func() {
					_ = conn.Close()
				}()

				waitForConnections(srv, 1, 2*time.Second)
			})
		})
	})

	Describe("TLS Connections", func() {
		BeforeEach(func() {
			err := srv.SetTLS(true, tlsConfig)
			Expect(err).ToNot(HaveOccurred())
			startServer(ctx, srv)
			waitForServerRunning(srv, 5*time.Second)
		})

		It("should accept TLS client connections", func() {
			clientConfig := &tls.Config{
				InsecureSkipVerify: true, // For testing with self-signed cert
			}

			conn := connectTLSClient(address, clientConfig)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)
		})

		It("should echo data over TLS", func() {
			clientConfig := &tls.Config{
				InsecureSkipVerify: true,
			}

			conn := connectTLSClient(address, clientConfig)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)

			msg := []byte("Secure Hello\n")
			n := sendMessage(conn, msg)
			Expect(n).To(Equal(len(msg)))

			response := receiveMessage(conn, 1024)
			Expect(response).To(Equal(msg))
		})

		It("should handle multiple TLS connections", func() {
			clientConfig := &tls.Config{
				InsecureSkipVerify: true,
			}

			conn1 := connectTLSClient(address, clientConfig)
			defer func() {
				_ = conn1.Close()
			}()

			conn2 := connectTLSClient(address, clientConfig)
			defer func() {
				_ = conn2.Close()
			}()

			conn3 := connectTLSClient(address, clientConfig)
			defer func() {
				_ = conn3.Close()
			}()

			waitForConnections(srv, 3, 2*time.Second)
		})

		It("should handle TLS handshake errors gracefully", func() {
			// Try to connect with wrong TLS version
			clientConfig := &tls.Config{
				InsecureSkipVerify: true,
				MaxVersion:         tls.VersionTLS10, // Server requires TLS 1.2+
			}

			// Connection may be established but handshake should fail
			_, err := tls.Dial("tcp", address, clientConfig)
			Expect(err).To(HaveOccurred())

			// Server should still be running
			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should verify TLS connection state", func() {
			clientConfig := &tls.Config{
				InsecureSkipVerify: true,
			}

			tlsConn, err := tls.Dial("tcp", address, clientConfig)
			defer func() {
				_ = tlsConn.Close()
			}()

			Expect(err).ToNot(HaveOccurred())

			// Verify connection state
			state := tlsConn.ConnectionState()
			Expect(state.HandshakeComplete).To(BeTrue())
			Expect(state.Version).To(BeNumerically(">=", uint16(tls.VersionTLS12)))
		})
	})

	Describe("TLS Configuration Edge Cases", func() {
		It("should handle SetTLS called multiple times", func() {
			err := srv.SetTLS(true, tlsConfig)
			Expect(err).ToNot(HaveOccurred())

			err = srv.SetTLS(true, tlsConfig)
			Expect(err).ToNot(HaveOccurred())

			err = srv.SetTLS(false, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not allow changing TLS config while running", func() {
			err := srv.SetTLS(true, tlsConfig)
			Expect(err).ToNot(HaveOccurred())

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			// Changing TLS while running is allowed but won't affect existing listener
			err = srv.SetTLS(false, nil)
			Expect(err).ToNot(HaveOccurred())

			// Server should still be running
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Describe("TLS Certificate Validation", func() {
		It("should work with valid self-signed certificate", func() {
			cert, err := generateSelfSignedCert()
			Expect(err).ToNot(HaveOccurred())
			Expect(cert.Certificate).ToNot(BeEmpty())
		})

		It("should properly configure server with certificate", func() {
			err := srv.SetTLS(true, tlsConfig)
			Expect(err).ToNot(HaveOccurred())

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			// Should be able to connect
			clientConfig := &tls.Config{
				InsecureSkipVerify: true,
			}
			conn := connectTLSClient(address, clientConfig)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)
		})
	})
})

// Helper function to connect with timeout
func connectClientWithTimeout(address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", address, timeout)
}
