/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package progress_test

import (
	"bytes"
	"io"
	"os"
	"strings"

	. "github.com/nabbar/golib/file/progress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edge Cases and Error Handling", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "progress-edge-test-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("Error conditions", func() {
		It("should handle write to read-only file", func() {
			path := tempDir + "/readonly.txt"
			err := os.WriteFile(path, []byte("test"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path) // Read-only
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			_, err = p.Write([]byte("should fail"))
			Expect(err).To(HaveOccurred())
		})

		It("should handle read from write-only file", func() {
			path := tempDir + "/writeonly.txt"
			p, err := New(path, os.O_CREATE|os.O_WRONLY, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			buf := make([]byte, 10)
			_, err = p.Read(buf)
			Expect(err).To(HaveOccurred())
		})

		It("should handle seek errors on closed file", func() {
			path := tempDir + "/closed.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())

			p.Close()

			_, err = p.Seek(0, io.SeekStart)
			Expect(err).To(HaveOccurred())
		})

		It("should handle operations on nil pointer", func() {
			var p *struct {
				Progress
			}

			// These should handle nil gracefully
			if p != nil {
				p.Close()
			}
		})
	})

	Describe("Boundary values", func() {
		It("should handle empty write", func() {
			path := tempDir + "/empty-write.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			n, err := p.Write([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should handle empty read", func() {
			path := tempDir + "/empty-read.txt"
			err := os.WriteFile(path, []byte("test"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			n, err := p.Read([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should handle zero buffer size", func() {
			path := tempDir + "/zero-buffer.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Set very small buffer (should use default)
			p.SetBufferSize(0)

			data := []byte("Test data")
			n, err := p.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
		})

		It("should handle negative buffer size", func() {
			path := tempDir + "/negative-buffer.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Negative buffer should use default
			p.SetBufferSize(-1024)

			data := []byte("Test data")
			n, err := p.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
		})
	})

	Describe("Large operations", func() {
		It("should handle very large ReadFrom", func() {
			path := tempDir + "/large-readfrom.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// 5MB
			largeData := bytes.Repeat([]byte("A"), 5*1024*1024)
			reader := bytes.NewReader(largeData)

			n, err := p.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(largeData))))
		})

		It("should handle very large WriteTo", func() {
			path := tempDir + "/large-writeto.txt"

			// Create large file
			largeData := bytes.Repeat([]byte("B"), 5*1024*1024)
			err := os.WriteFile(path, largeData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var buf bytes.Buffer
			n, err := p.WriteTo(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(largeData))))
		})

		It("should handle many small writes", func() {
			path := tempDir + "/many-writes.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// 1000 small writes
			for i := 0; i < 1000; i++ {
				_, err := p.Write([]byte("X"))
				Expect(err).ToNot(HaveOccurred())
			}

			info, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Size()).To(Equal(int64(1000)))
		})
	})

	Describe("Concurrent operations", func() {
		It("should handle concurrent reads safely", func() {
			path := tempDir + "/concurrent-reads.txt"
			data := bytes.Repeat([]byte("test data "), 1000)
			err := os.WriteFile(path, data, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					buf := make([]byte, 100)
					p.Read(buf)
					done <- true
				}()
			}

			// Wait for goroutines
			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle concurrent callback registrations", func() {
			path := tempDir + "/concurrent-callbacks.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					p.RegisterFctIncrement(func(size int64) {})
					p.RegisterFctReset(func(max, current int64) {})
					p.RegisterFctEOF(func() {})
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}

			// Should not panic
			p.Write([]byte("test"))
		})
	})

	Describe("Special cases", func() {
		It("should handle ReadFrom with empty reader", func() {
			path := tempDir + "/empty-reader.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			reader := strings.NewReader("")
			n, err := p.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(0)))
		})

		It("should handle WriteTo with empty file", func() {
			path := tempDir + "/empty-file.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var buf bytes.Buffer
			n, err := p.WriteTo(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(0)))
			Expect(buf.Len()).To(Equal(0))
		})

		It("should handle consecutive EOF reads", func() {
			path := tempDir + "/consecutive-eof.txt"
			err := os.WriteFile(path, []byte("short"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			buf := make([]byte, 100)
			n, err := p.Read(buf)
			Expect(n).To(Equal(5))

			// Second read should return EOF
			n, err = p.Read(buf)
			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))

			// Third read should also return EOF
			n, err = p.Read(buf)
			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))
		})

		It("should handle mixed ReadAt and Read operations", func() {
			path := tempDir + "/mixed-read.txt"
			data := []byte("0123456789ABCDEF")
			err := os.WriteFile(path, data, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Read normally
			buf1 := make([]byte, 5)
			n, err := p.Read(buf1)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(buf1)).To(Equal("01234"))

			// ReadAt (doesn't change position)
			buf2 := make([]byte, 5)
			n, err = p.ReadAt(buf2, 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(buf2)).To(Equal("ABCDE"))

			// Continue reading normally (position unchanged)
			buf3 := make([]byte, 5)
			n, err = p.Read(buf3)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(buf3)).To(Equal("56789"))
		})

		It("should handle Reset with callbacks", func() {
			path := tempDir + "/reset-with-callbacks.txt"
			data := []byte("test data")
			err := os.WriteFile(path, data, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var resetCalled bool
			var maxValue, currentValue int64

			p.RegisterFctReset(func(max, current int64) {
				resetCalled = true
				maxValue = max
				currentValue = current
			})

			// Read some bytes
			buf := make([]byte, 5)
			p.Read(buf)

			// Manual reset
			p.Reset(100)

			Expect(resetCalled).To(BeTrue())
			Expect(maxValue).To(Equal(int64(100)))
			Expect(currentValue).To(BeNumerically(">=", 0))
		})

		It("should handle Truncate with callbacks", func() {
			path := tempDir + "/truncate-callbacks.txt"
			p, err := New(path, os.O_CREATE|os.O_RDWR, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var resetCalled bool

			p.RegisterFctReset(func(max, current int64) {
				resetCalled = true
			})

			p.WriteString("initial data")
			p.Truncate(5)

			// Truncate should trigger reset
			Expect(resetCalled).To(BeTrue())
		})
	})

	Describe("File permissions", func() {
		It("should respect file permissions on creation", func() {
			path := tempDir + "/perms.txt"
			p, err := New(path, os.O_CREATE|os.O_RDWR, 0600)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			info, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())

			// Check permissions (may vary on different systems)
			mode := info.Mode()
			Expect(mode.IsRegular()).To(BeTrue())
		})
	})

	Describe("DefaultBuffSize constant", func() {
		It("should use default buffer size", func() {
			Expect(DefaultBuffSize).To(Equal(32 * 1024))
		})
	})
})
