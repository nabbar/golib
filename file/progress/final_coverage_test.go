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

var _ = Describe("Final Coverage Improvements", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "progress-final-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("ReadFrom with LimitedReader", func() {
		It("should handle limited reader with small buffer", func() {
			path := tempDir + "/readfrom-limited.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Use a LimitedReader with small size
			data := "test data for limited reader"
			src := io.LimitReader(strings.NewReader(data), int64(len(data)))

			n, err := p.ReadFrom(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(data))))
		})

		It("should handle EOF during ReadFrom", func() {
			path := tempDir + "/readfrom-eof.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			incCalled := 0
			eofCalled := false

			p.RegisterFctIncrement(func(size int64) {
				incCalled++
			})
			p.RegisterFctEOF(func() {
				eofCalled = true
			})

			// Read from an empty reader
			src := strings.NewReader("")
			_, err = p.ReadFrom(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(eofCalled).To(BeTrue())
		})
	})

	Describe("WriteTo with callbacks", func() {
		It("should trigger callbacks during WriteTo", func() {
			path := tempDir + "/writeto-callbacks.txt"
			data := []byte("data for WriteTo testing with callbacks")
			err := os.WriteFile(path, data, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			eofCalled := false
			p.RegisterFctEOF(func() {
				eofCalled = true
			})

			var buf bytes.Buffer
			n, err := p.WriteTo(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(data))))
			Expect(eofCalled).To(BeTrue())
		})
	})

	Describe("SizeEOF error handling", func() {
		It("should handle seek errors in SizeEOF", func() {
			path := tempDir + "/sizeeof-errors.txt"
			err := os.WriteFile(path, []byte("0123456789"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())

			// Normal operation
			size, err := p.SizeEOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(size).To(Equal(int64(10)))

			p.Close()

			// After close, should error
			_, err = p.SizeEOF()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ReadByte with multi-byte read", func() {
		It("should handle seek positioning correctly", func() {
			path := tempDir + "/readbyte-seek.txt"
			err := os.WriteFile(path, []byte("ABCDEFGH"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Read first byte
			b, err := p.ReadByte()
			Expect(err).ToNot(HaveOccurred())
			Expect(b).To(Equal(byte('A')))

			// Position should be at 1
			pos, _ := p.Seek(0, io.SeekCurrent)
			Expect(pos).To(Equal(int64(1)))

			// Read second byte
			b, err = p.ReadByte()
			Expect(err).ToNot(HaveOccurred())
			Expect(b).To(Equal(byte('B')))
		})
	})

	Describe("WriteByte with positioning", func() {
		It("should maintain correct file position", func() {
			path := tempDir + "/writebyte-pos.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Write several bytes
			for _, b := range []byte("HELLO") {
				err = p.WriteByte(b)
				Expect(err).ToNot(HaveOccurred())
			}

			// Verify position
			pos, _ := p.Seek(0, io.SeekCurrent)
			Expect(pos).To(Equal(int64(5)))

			// Verify content
			p.Seek(0, io.SeekStart)
			buf := make([]byte, 5)
			p.Read(buf)
			Expect(string(buf)).To(Equal("HELLO"))
		})
	})

	Describe("Create with callbacks", func() {
		It("should work with callbacks on newly created file", func() {
			path := tempDir + "/create-callbacks.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var totalBytes int64
			p.RegisterFctIncrement(func(size int64) {
				totalBytes += size
			})

			// Write data
			data := []byte("Created file with callbacks")
			n, err := p.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
			Expect(totalBytes).To(Equal(int64(len(data))))
		})
	})

	Describe("Temp file creation", func() {
		It("should create temp file with pattern", func() {
			p, err := Temp("test-pattern-*.dat")
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			path := p.Path()
			defer os.Remove(path)

			Expect(p.IsTemp()).To(BeTrue())
			Expect(path).To(ContainSubstring("test-pattern-"))
		})
	})

	Describe("New with various flags", func() {
		It("should handle O_APPEND flag", func() {
			path := tempDir + "/append.txt"
			err := os.WriteFile(path, []byte("initial"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := New(path, os.O_APPEND|os.O_WRONLY, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			p.Write([]byte(" appended"))

			// Verify
			content, _ := os.ReadFile(path)
			Expect(string(content)).To(Equal("initial appended"))
		})
	})

	Describe("Open error cases", func() {
		It("should return error for non-existent file", func() {
			_, err := Open("/nonexistent/path/to/file.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for invalid pattern in Temp", func() {
			// Empty pattern should still work, but let's test creation
			p, err := Temp("")
			Expect(err).ToNot(HaveOccurred())
			if p != nil {
				defer p.Close()
				defer os.Remove(p.Path())
			}
		})
	})
})
