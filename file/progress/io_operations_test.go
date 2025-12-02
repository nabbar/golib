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

var _ = Describe("IO Operations", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "progress-io-test-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("Write operations", func() {
		It("should write data to file", func() {
			path := tempDir + "/write-test.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			data := []byte("Hello, World!")
			n, err := p.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))

			// Verify content
			content, err := os.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(Equal(data))
		})

		It("should write multiple times", func() {
			path := tempDir + "/write-multiple.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			data1 := []byte("First ")
			data2 := []byte("Second ")
			data3 := []byte("Third")

			n1, err := p.Write(data1)
			Expect(err).ToNot(HaveOccurred())
			Expect(n1).To(Equal(len(data1)))

			n2, err := p.Write(data2)
			Expect(err).ToNot(HaveOccurred())
			Expect(n2).To(Equal(len(data2)))

			n3, err := p.Write(data3)
			Expect(err).ToNot(HaveOccurred())
			Expect(n3).To(Equal(len(data3)))

			// Verify total content
			content, err := os.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal("First Second Third"))
		})

		It("should write string", func() {
			path := tempDir + "/write-string.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			text := "String content"
			n, err := p.WriteString(text)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(text)))

			content, err := os.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal(text))
		})

		It("should write at specific offset", func() {
			path := tempDir + "/write-at.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Write initial content
			p.WriteString("0123456789")

			// Write at offset 5
			data := []byte("ABCDE")
			n, err := p.WriteAt(data, 5)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))

			// Verify content
			content, err := os.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal("01234ABCDE"))
		})
	})

	Describe("Read operations", func() {
		It("should read data from file", func() {
			path := tempDir + "/read-test.txt"
			testData := []byte("Read this content")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			buf := make([]byte, len(testData))
			n, err := p.Read(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(testData)))
			Expect(buf).To(Equal(testData))
		})

		It("should read in chunks", func() {
			path := tempDir + "/read-chunks.txt"
			testData := []byte("This is a longer text for chunk reading")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var result []byte
			buf := make([]byte, 10)
			for {
				n, err := p.Read(buf)
				if n > 0 {
					result = append(result, buf[:n]...)
				}
				if err == io.EOF {
					break
				}
				Expect(err).ToNot(HaveOccurred())
			}

			Expect(result).To(Equal(testData))
		})

		It("should read at specific offset", func() {
			path := tempDir + "/read-at.txt"
			testData := []byte("0123456789ABCDEF")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			buf := make([]byte, 5)
			n, err := p.ReadAt(buf, 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(buf)).To(Equal("ABCDE"))
		})

		It("should handle EOF", func() {
			path := tempDir + "/read-eof.txt"
			testData := []byte("Short")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			buf := make([]byte, 100)
			n, err := p.Read(buf)
			Expect(n).To(Equal(len(testData)))
			Expect(err).ToNot(HaveOccurred())

			// Second read should return EOF
			n, err = p.Read(buf)
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))
		})
	})

	Describe("Seek operations", func() {
		It("should seek to beginning", func() {
			path := tempDir + "/seek-start.txt"
			testData := []byte("0123456789")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Read 5 bytes
			buf := make([]byte, 5)
			p.Read(buf)

			// Seek back to start
			pos, err := p.Seek(0, io.SeekStart)
			Expect(err).ToNot(HaveOccurred())
			Expect(pos).To(Equal(int64(0)))

			// Read again from start
			n, err := p.Read(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(buf)).To(Equal("01234"))
		})

		It("should seek to end", func() {
			path := tempDir + "/seek-end.txt"
			testData := []byte("0123456789")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			pos, err := p.Seek(0, io.SeekEnd)
			Expect(err).ToNot(HaveOccurred())
			Expect(pos).To(Equal(int64(len(testData))))
		})

		It("should seek relative to current position", func() {
			path := tempDir + "/seek-current.txt"
			testData := []byte("0123456789ABCDEF")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Read 5 bytes (position is now 5)
			buf := make([]byte, 5)
			p.Read(buf)

			// Seek forward 3 bytes from current
			pos, err := p.Seek(3, io.SeekCurrent)
			Expect(err).ToNot(HaveOccurred())
			Expect(pos).To(Equal(int64(8)))

			// Read and verify
			n, err := p.Read(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(buf)).To(Equal("89ABC"))
		})
	})

	Describe("ReadFrom and WriteTo", func() {
		It("should read from reader", func() {
			path := tempDir + "/read-from.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			sourceData := []byte("Data from reader")
			reader := bytes.NewReader(sourceData)

			n, err := p.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(sourceData))))

			// Verify content
			content, err := os.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(Equal(sourceData))
		})

		It("should write to writer", func() {
			path := tempDir + "/write-to.txt"
			testData := []byte("Data to write")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var buf bytes.Buffer
			n, err := p.WriteTo(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(testData))))
			Expect(buf.Bytes()).To(Equal(testData))
		})

		It("should handle large ReadFrom", func() {
			path := tempDir + "/read-from-large.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Create large data (1MB)
			largeData := bytes.Repeat([]byte("A"), 1024*1024)
			reader := bytes.NewReader(largeData)

			n, err := p.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(largeData))))
		})

		It("should handle io.LimitedReader", func() {
			path := tempDir + "/limited-reader.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			data := []byte("0123456789ABCDEF")
			reader := &io.LimitedReader{
				R: bytes.NewReader(data),
				N: 10, // Limit to 10 bytes
			}

			n, err := p.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(10)))

			// Verify only 10 bytes were written
			content, err := os.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(content)).To(Equal(10))
			Expect(string(content)).To(Equal("0123456789"))
		})
	})

	Describe("Buffer size management", func() {
		It("should use default buffer size", func() {
			path := tempDir + "/default-buffer.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Write large data to test buffering
			largeData := strings.Repeat("X", 100000)
			reader := strings.NewReader(largeData)

			n, err := p.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(largeData))))
		})

		It("should use custom buffer size", func() {
			path := tempDir + "/custom-buffer.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Set custom buffer size
			p.SetBufferSize(8192)

			largeData := strings.Repeat("Y", 50000)
			reader := strings.NewReader(largeData)

			n, err := p.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(largeData))))
		})

		It("should handle small buffer size", func() {
			path := tempDir + "/small-buffer.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Set very small buffer (will use default instead)
			p.SetBufferSize(512)

			data := []byte("Test with small buffer")
			reader := bytes.NewReader(data)

			n, err := p.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(data))))
		})
	})
})
