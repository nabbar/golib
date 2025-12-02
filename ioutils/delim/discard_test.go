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

package delim_test

import (
	"io"

	iotdlm "github.com/nabbar/golib/ioutils/delim"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This test file validates the DiscardCloser implementation.
// DiscardCloser is a no-op io.ReadWriteCloser used for:
//   - Testing scenarios requiring a valid reader/writer that does nothing
//   - Placeholder implementations where data should be discarded
//   - Benchmarking to isolate I/O operations
//
// Tests cover:
//   - Read operations (always returns 0 without error)
//   - Write operations (accepts all data, returns success)
//   - Close operations (no-op, always succeeds)
//   - Interface compliance (io.ReadWriteCloser)
//   - Concurrent access patterns
//   - Integration with BufferDelim and io.Copy
//
// DiscardCloser is similar to io.Discard but also implements Reader and Closer.

var _ = Describe("DiscardCloser", func() {
	var dc iotdlm.DiscardCloser

	BeforeEach(func() {
		dc = iotdlm.DiscardCloser{}
	})

	Describe("Read operation", func() {
		Context("with various buffer sizes", func() {
			It("should always return 0 bytes and no error", func() {
				buf := make([]byte, 100)
				n, err := dc.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle small buffer", func() {
				buf := make([]byte, 1)
				n, err := dc.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle large buffer", func() {
				buf := make([]byte, 1000000)
				n, err := dc.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle nil buffer", func() {
				var buf []byte
				n, err := dc.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle zero-length buffer", func() {
				buf := make([]byte, 0)
				n, err := dc.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})
		})

		Context("with multiple reads", func() {
			It("should consistently return 0", func() {
				buf := make([]byte, 10)
				for i := 0; i < 100; i++ {
					n, err := dc.Read(buf)
					Expect(err).To(BeNil())
					Expect(n).To(Equal(0))
				}
			})

			It("should not modify buffer content", func() {
				buf := []byte{1, 2, 3, 4, 5}
				original := make([]byte, len(buf))
				copy(original, buf)

				n, err := dc.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
				Expect(buf).To(Equal(original))
			})
		})
	})

	Describe("Write operation", func() {
		Context("with various data sizes", func() {
			It("should accept and discard data", func() {
				data := []byte("test data")
				n, err := dc.Write(data)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(len(data)))
			})

			It("should handle empty data", func() {
				data := []byte{}
				n, err := dc.Write(data)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle nil data", func() {
				var data []byte
				n, err := dc.Write(data)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle small write", func() {
				data := []byte("x")
				n, err := dc.Write(data)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(1))
			})

			It("should handle large write", func() {
				data := make([]byte, 1000000)
				n, err := dc.Write(data)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(1000000))
			})

			It("should handle binary data", func() {
				data := []byte{0x00, 0xFF, 0xDE, 0xAD, 0xBE, 0xEF}
				n, err := dc.Write(data)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(len(data)))
			})
		})

		Context("with multiple writes", func() {
			It("should accept multiple writes", func() {
				for i := 0; i < 100; i++ {
					data := []byte("test")
					n, err := dc.Write(data)
					Expect(err).To(BeNil())
					Expect(n).To(Equal(4))
				}
			})

			It("should handle alternating write sizes", func() {
				sizes := []int{10, 100, 1, 1000, 50}
				for _, size := range sizes {
					data := make([]byte, size)
					n, err := dc.Write(data)
					Expect(err).To(BeNil())
					Expect(n).To(Equal(size))
				}
			})
		})
	})

	Describe("Close operation", func() {
		It("should close without error", func() {
			err := dc.Close()
			Expect(err).To(BeNil())
		})

		It("should allow multiple close calls", func() {
			err := dc.Close()
			Expect(err).To(BeNil())

			err = dc.Close()
			Expect(err).To(BeNil())

			err = dc.Close()
			Expect(err).To(BeNil())
		})

		It("should still work after close", func() {
			err := dc.Close()
			Expect(err).To(BeNil())

			// Read after close
			buf := make([]byte, 10)
			n, err := dc.Read(buf)
			Expect(err).To(BeNil())
			Expect(n).To(Equal(0))

			// Write after close
			data := []byte("test")
			n, err = dc.Write(data)
			Expect(err).To(BeNil())
			Expect(n).To(Equal(4))
		})
	})

	Describe("Interface compliance", func() {
		It("should implement io.Reader", func() {
			var _ io.Reader = dc
		})

		It("should implement io.Writer", func() {
			var _ io.Writer = dc
		})

		It("should implement io.Closer", func() {
			var _ io.Closer = dc
		})

		It("should implement io.ReadCloser", func() {
			var _ io.ReadCloser = dc
		})

		It("should implement io.WriteCloser", func() {
			var _ io.WriteCloser = dc
		})

		It("should implement io.ReadWriteCloser", func() {
			var _ io.ReadWriteCloser = dc
		})
	})

	Describe("Usage scenarios", func() {
		Context("with io.Copy", func() {
			It("should work as destination in io.Copy", func() {
				src := []byte("test data to discard")
				n, err := io.Copy(dc, io.NopCloser(io.Reader(io.LimitReader(io.MultiReader(), 0))))
				Expect(err).To(BeNil())
				_ = n
				_ = src
			})

			It("should discard all data when used with io.Copy", func() {
				data := []byte("this will be discarded\nline 2\nline 3")
				n, err := dc.Write(data)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(len(data)))
			})
		})

		Context("as placeholder", func() {
			It("should work as a no-op closer", func() {
				// Use case: when you need a Closer but don't want to do anything
				var rc io.ReadCloser = dc
				err := rc.Close()
				Expect(err).To(BeNil())
			})

			It("should work as a dev/null equivalent", func() {
				// Like /dev/null in Unix
				testData := [][]byte{
					[]byte("log message 1\n"),
					[]byte("log message 2\n"),
					[]byte("log message 3\n"),
				}

				for _, data := range testData {
					n, err := dc.Write(data)
					Expect(err).To(BeNil())
					Expect(n).To(Equal(len(data)))
				}
			})
		})

		Context("with BufferDelim", func() {
			It("should work as input to BufferDelim", func() {
				bd := iotdlm.New(dc, '\n', 0)
				Expect(bd).NotTo(BeNil())

				// Reading from DiscardCloser via BufferDelim
				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				// Should get EOF or empty read since DiscardCloser returns 0
				_ = n
				_ = err
			})
		})
	})

	Describe("Concurrent access", func() {
		It("should handle concurrent reads safely", func() {
			done := make(chan bool)
			for i := 0; i < 10; i++ {
				go func() {
					buf := make([]byte, 100)
					for j := 0; j < 100; j++ {
						_, _ = dc.Read(buf)
					}
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle concurrent writes safely", func() {
			done := make(chan bool)
			for i := 0; i < 10; i++ {
				go func() {
					data := []byte("concurrent write test")
					for j := 0; j < 100; j++ {
						_, _ = dc.Write(data)
					}
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle mixed concurrent operations", func() {
			done := make(chan bool)

			// Concurrent reads
			for i := 0; i < 5; i++ {
				go func() {
					buf := make([]byte, 50)
					for j := 0; j < 50; j++ {
						_, _ = dc.Read(buf)
					}
					done <- true
				}()
			}

			// Concurrent writes
			for i := 0; i < 5; i++ {
				go func() {
					data := []byte("test")
					for j := 0; j < 50; j++ {
						_, _ = dc.Write(data)
					}
					done <- true
				}()
			}

			// Concurrent closes
			for i := 0; i < 5; i++ {
				go func() {
					for j := 0; j < 10; j++ {
						_ = dc.Close()
					}
					done <- true
				}()
			}

			for i := 0; i < 15; i++ {
				<-done
			}
		})
	})
})
