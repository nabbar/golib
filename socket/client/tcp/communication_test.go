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
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	sckclt "github.com/nabbar/golib/socket/client/tcp"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Client Communication", func() {
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

	Describe("Write", func() {
		Context("with connected client", func() {
			It("should write data successfully", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()
				connectClient(ctx, cli)

				msg := []byte("Hello, World!\n")
				n, err := cli.Write(msg)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(msg)))
			})

			It("should write multiple messages", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()
				connectClient(ctx, cli)

				for i := 0; i < 10; i++ {
					msg := []byte("Message " + string(rune('0'+i)) + "\n")
					n, err := cli.Write(msg)
					Expect(err).ToNot(HaveOccurred())
					Expect(n).To(Equal(len(msg)))
				}
			})

			It("should handle empty write", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()
				connectClient(ctx, cli)

				n, err := cli.Write([]byte{})
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(0))
			})

			It("should handle nil write", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()
				connectClient(ctx, cli)

				n, err := cli.Write(nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(0))
			})
		})

		Context("without connection", func() {
			It("should fail to write", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()

				msg := []byte("Hello\n")
				n, err := cli.Write(msg)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(sckclt.ErrConnection))
				Expect(n).To(Equal(0))
			})
		})
	})

	Describe("Read", func() {
		Context("with connected client", func() {
			It("should read echoed data", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()
				connectClient(ctx, cli)

				msg := []byte("Hello, World!\n")
				n, err := cli.Write(msg)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(msg)))

				response := make([]byte, len(msg))
				n, err = io.ReadFull(cli, response)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(msg)))
				Expect(response).To(Equal(msg))
			})

			It("should read multiple messages", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()
				connectClient(ctx, cli)

				for i := 0; i < 10; i++ {
					msg := []byte("Message " + string(rune('0'+i)) + "\n")
					response := sendAndReceive(cli, msg)
					Expect(response).To(Equal(msg))
				}
			})

			It("should handle partial reads", func() {
				cli := createClient(address)
				defer func() {
					_ = cli.Close()
				}()
				connectClient(ctx, cli)

				msg := []byte("Hello, World!\n")
				n, err := cli.Write(msg)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(msg)))

				// Read in chunks
				chunk := make([]byte, 5)
				total := 0
				for total < len(msg) {
					n, err := cli.Read(chunk)
					if err != nil {
						break
					}
					total += n
				}
				Expect(total).To(Equal(len(msg)))
			})
		})

		Context("without connection", func() {
			It("should fail to read", func() {
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
		})
	})

	Describe("Once", func() {
		Context("with valid request", func() {
			It("should send and receive data", func() {
				cli := createClient(address)
				// Don't defer close, Once will close it

				msg := []byte("Hello from Once!\n")
				request := bytes.NewReader(msg)

				var response bytes.Buffer
				err := cli.Once(ctx, request, func(r io.Reader) {
					_, _ = io.Copy(&response, r)
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(response.Bytes()).To(Equal(msg))
			})

			It("should handle multiple calls with different clients", func() {
				wg := sync.WaitGroup{}
				for i := 0; i < 5; i++ {
					fct := func() {
						cli := createClient(address)
						defer func() {
							_ = cli.Close()
						}()

						msg := []byte("Message " + string(rune('0'+i)) + "\n")
						request := bytes.NewReader(msg)

						var response bytes.Buffer
						err := cli.Once(ctx, request, func(r io.Reader) {
							_, e := io.Copy(&response, r)
							Expect(e).ToNot(HaveOccurred())
						})

						Expect(err).ToNot(HaveOccurred())
						Expect(response.Bytes()).To(Equal(msg))
					}
					wg.Go(fct)
				}
				wg.Wait()
			})

			It("should work with nil response callback", func() {
				cli := createClient(address)

				msg := []byte("Hello!\n")
				request := bytes.NewReader(msg)

				err := cli.Once(ctx, request, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle empty request", func() {
				cli := createClient(address)

				request := bytes.NewReader([]byte{})

				err := cli.Once(ctx, request, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with connection issues", func() {
			It("should fail when server is not available", func() {
				noServerAddr := getTestAddress()
				cli := createClient(noServerAddr)

				msg := []byte("Hello!\n")
				request := bytes.NewReader(msg)

				err := cli.Once(ctx, request, nil)
				Expect(err).To(HaveOccurred())
			})

			It("should handle context cancellation", func() {
				cli := createClient(address)

				cancelCtx, cancelFunc := context.WithCancel(ctx)
				cancelFunc() // Cancel immediately

				msg := []byte("Hello!\n")
				request := bytes.NewReader(msg)

				err := cli.Once(cancelCtx, request, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with response processing", func() {
			It("should process response in callback", func() {
				cli := createClient(address)

				msg := []byte("Test message\n")
				request := bytes.NewReader(msg)

				callbackCalled := false
				var receivedData []byte

				err := cli.Once(ctx, request, func(r io.Reader) {
					callbackCalled = true
					data, _ := io.ReadAll(r)
					receivedData = data
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(callbackCalled).To(BeTrue())
				Expect(receivedData).To(Equal(msg))
			})

			It("should handle streaming response", func() {
				cli := createClient(address)

				// Send multiple lines
				msgs := []string{"Line 1\n", "Line 2\n", "Line 3\n"}
				combined := strings.Join(msgs, "")
				request := bytes.NewReader([]byte(combined))

				lines := make([]string, 0)
				err := cli.Once(ctx, request, func(r io.Reader) {
					buf := make([]byte, 1024)
					for {
						n, err := r.Read(buf)
						if n > 0 {
							lines = append(lines, string(buf[:n]))
						}
						if err != nil {
							break
						}
					}
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(strings.Join(lines, "")).To(Equal(combined))
			})
		})
	})

	Describe("Bidirectional Communication", func() {
		It("should handle concurrent read and write", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()
			connectClient(ctx, cli)

			done := make(chan bool)
			errors := make([]error, 0)

			// Writer goroutine
			go func() {
				defer GinkgoRecover()
				for i := 0; i < 5; i++ {
					msg := []byte("Message " + string(rune('0'+i)) + "\n")
					_, err := cli.Write(msg)
					if err != nil {
						errors = append(errors, err)
					}
					time.Sleep(10 * time.Millisecond)
				}
			}()

			// Reader goroutine
			go func() {
				defer GinkgoRecover()
				defer close(done)
				buf := make([]byte, 1024)
				count := 0
				for count < 5 {
					n, err := cli.Read(buf)
					if err != nil {
						if err != io.EOF {
							errors = append(errors, err)
						}
						return
					}
					if n > 0 {
						count++
					}
				}
			}()

			select {
			case <-done:
				Expect(errors).To(BeEmpty())
			case <-time.After(5 * time.Second):
				Fail("Timeout waiting for concurrent communication")
			}
		})

		It("should handle rapid message exchange", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()
			connectClient(ctx, cli)

			count := 100
			for i := 0; i < count; i++ {
				msg := []byte("Fast message\n")
				response := sendAndReceive(cli, msg)
				Expect(response).To(Equal(msg))
			}
		})
	})

	Describe("Performance", func() {
		It("should handle high throughput", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()
			connectClient(ctx, cli)

			iterations := 1000
			msgSize := 1024
			msg := make([]byte, msgSize)
			for i := range msg {
				msg[i] = byte(i % 256)
			}

			start := time.Now()
			for i := 0; i < iterations; i++ {
				n, err := cli.Write(msg)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(msg)))

				response := make([]byte, len(msg))
				_, err = io.ReadFull(cli, response)
				Expect(err).ToNot(HaveOccurred())
			}
			elapsed := time.Since(start)

			// Should complete in reasonable time (allow 10 seconds for 1000 iterations)
			Expect(elapsed).To(BeNumerically("<", 10*time.Second))
		})

		It("should maintain performance with counting handler", func() {
			defer GinkgoRecover()

			counter := new(atomic.Int32)

			counterAddr := getTestAddress()
			counterSrv := createAndRegisterServer(counterAddr, countingHandler(counter))
			defer func() {
				_ = counterSrv.Shutdown(ctx)
			}()

			startServer(ctx, counterSrv)
			waitForServerRunning(counterAddr, 5*time.Second)

			iterations := 25
			msg := []byte("Count me\n")

			wg := sync.WaitGroup{}
			for i := 0; i < iterations; i++ {
				wg.Go(func() {
					defer GinkgoRecover()

					cli := createClient(counterAddr)
					defer func() {
						_ = cli.Close()
					}()

					connectClient(ctx, cli)

					rsp := bytes.NewBuffer(make([]byte, 0, 2*len(msg)))

					n, e := io.Copy(cli, bytes.NewReader(msg))
					Expect(e).ToNot(HaveOccurred())
					Expect(n).To(Equal(int64(len(msg))))

					n, e = io.Copy(rsp, cli)
					Expect(e).ToNot(HaveOccurred())
					Expect(n).To(Equal(int64(len(msg))))
				})
			}
			wg.Wait()

			// Verify handler was called correct number of times
			// iteraction + 1 connection for waiting server up
			Eventually(counter.Load, 5*time.Second, 50*time.Millisecond).Should(Equal(int32(iterations) + 1))
		})
	})
})
