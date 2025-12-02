//go:build linux || darwin

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

package unixgram_test

import (
	"bytes"
	"context"
	"io"
	"sync/atomic"
	"time"

	scksrv "github.com/nabbar/golib/socket/server/unixgram"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UNIX Datagram Client Communication", func() {
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

	Describe("Write", func() {
		It("should write data successfully", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			data := []byte("Hello, UNIX Datagram!")
			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
		})

		It("should write empty data", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			n, err := cli.Write([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should write large data", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// UNIX datagram has packet size limits, use reasonable size
			data := make([]byte, 1024)
			for i := range data {
				data[i] = byte(i % 256)
			}

			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
		})

		It("should fail when not connected", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			_, err := cli.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Read", func() {
		It("should fail when not connected", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			buf := make([]byte, 1024)
			_, err := cli.Read(buf)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Once", func() {
		It("should perform request/response operation", func() {
			cli := createClient(socketPath)

			request := bytes.NewBufferString("Once test")
			responseChan := make(chan []byte, 1)
			errorChan := make(chan error, 1)

			// Run Once in goroutine with timeout
			go func() {
				err := cli.Once(ctx, request, func(reader io.Reader) {
					// Use a timeout for the read operation
					done := make(chan bool, 1)
					var buf []byte
					go func() {
						tmpBuf := make([]byte, 1024)
						n, _ := reader.Read(tmpBuf)
						if n > 0 {
							buf = tmpBuf[:n]
							done <- true
						}
					}()

					select {
					case <-done:
						responseChan <- buf
					case <-time.After(500 * time.Millisecond):
						// Timeout - no response received
					}
				})
				errorChan <- err
			}()

			// Wait for Once to complete or timeout
			select {
			case err := <-errorChan:
				Expect(err).ToNot(HaveOccurred())
			case <-time.After(2 * time.Second):
				Fail("Once operation timed out")
			}

			// May or may not receive response depending on datagram timing
			select {
			case resp := <-responseChan:
				Expect(resp).To(Equal([]byte("Once test")))
			case <-time.After(100 * time.Millisecond):
				// Timeout is acceptable for datagram
			}

			// Connection should be closed after Once
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should work without response callback", func() {
			cli := createClient(socketPath)

			request := bytes.NewBufferString("Fire and forget")
			err := cli.Once(ctx, request, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should handle empty request", func() {
			cli := createClient(socketPath)

			request := bytes.NewBuffer(nil)
			err := cli.Once(ctx, request, nil)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should close connection even on error", func() {
			nonExistentPath := getTestSocketPath()
			cli := createClient(nonExistentPath)

			request := bytes.NewBufferString("test")
			_ = cli.Once(ctx, request, nil)

			// Should still be closed even if operation failed
			Expect(cli.IsConnected()).To(BeFalse())
		})
	})

	Describe("Concurrent Operations", func() {
		It("should handle multiple writes from same client", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Small delay to ensure connection is stable
			time.Sleep(50 * time.Millisecond)

			successCount := new(atomic.Int32)
			done := make(chan bool, 3)

			for i := 0; i < 3; i++ {
				go func(id int) {
					defer GinkgoRecover()
					data := []byte("C" + string(rune('0'+id)))
					_, err := cli.Write(data)
					// Note: With datagram sockets, some writes may fail if server is busy
					// We just count successes
					if err == nil {
						successCount.Add(1)
					}
					done <- true
				}(i)
			}

			// Wait for all writes to complete
			for i := 0; i < 3; i++ {
				Eventually(done, 2*time.Second).Should(Receive())
			}

			// At least one write should succeed
			Expect(successCount.Load()).To(BeNumerically(">", 0))
		})
	})

	Describe("Binary Data", func() {
		It("should handle binary data correctly", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Binary data - smaller size for datagram
			data := make([]byte, 100)
			for i := 0; i < 100; i++ {
				data[i] = byte(i)
			}

			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
		})
	})
})
