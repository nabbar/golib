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
	"io"
	"os"

	. "github.com/nabbar/golib/file/progress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Additional Coverage Tests", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "progress-coverage-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("ReadByte", func() {
		It("should read a single byte successfully", func() {
			path := tempDir + "/readbyte.txt"
			err := os.WriteFile(path, []byte("ABCD"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			b, err := p.ReadByte()
			Expect(err).ToNot(HaveOccurred())
			Expect(b).To(Equal(byte('A')))
		})

		It("should return EOF when reading past end", func() {
			path := tempDir + "/readbyte-eof.txt"
			err := os.WriteFile(path, []byte(""), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			_, err = p.ReadByte()
			Expect(err).To(Equal(io.EOF))
		})
	})

	Describe("WriteByte", func() {
		It("should write a single byte successfully", func() {
			path := tempDir + "/writebyte.txt"
			p, err := New(path, os.O_CREATE|os.O_RDWR, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			err = p.WriteByte('X')
			Expect(err).ToNot(HaveOccurred())

			// Verify
			p.Seek(0, io.SeekStart)
			b, err := p.ReadByte()
			Expect(err).ToNot(HaveOccurred())
			Expect(b).To(Equal(byte('X')))
		})
	})

	Describe("Temp file operations", func() {
		It("should mark temp file correctly", func() {
			// Use Temp which creates temporary files
			p, err := Temp("temp-*.txt")
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Verify it's marked as temp
			Expect(p.IsTemp()).To(BeTrue())

			path := p.Path()
			p.Write([]byte("test"))

			// Close temp file
			err = p.Close()
			Expect(err).ToNot(HaveOccurred())

			// Clean up manually if needed
			os.Remove(path)
		})
	})

	Describe("SizeEOF edge cases", func() {
		It("should calculate EOF size correctly at different positions", func() {
			path := tempDir + "/sizeeof.txt"
			err := os.WriteFile(path, []byte("0123456789"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// At start
			eof, err := p.SizeEOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(eof).To(Equal(int64(10)))

			// After reading
			buf := make([]byte, 3)
			p.Read(buf)

			eof, err = p.SizeEOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(eof).To(Equal(int64(7)))
		})
	})

	Describe("SizeBOF edge cases", func() {
		It("should return correct BOF size", func() {
			path := tempDir + "/sizebof.txt"
			err := os.WriteFile(path, []byte("0123456789"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// At start
			bof, err := p.SizeBOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(bof).To(Equal(int64(0)))

			// After reading
			buf := make([]byte, 7)
			p.Read(buf)

			bof, err = p.SizeBOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(bof).To(Equal(int64(7)))
		})
	})

	Describe("Sync operations", func() {
		It("should sync without error", func() {
			path := tempDir + "/sync.txt"
			p, err := New(path, os.O_CREATE|os.O_RDWR, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			p.Write([]byte("data"))
			err = p.Sync()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("ReadAt and WriteAt", func() {
		It("should read at specific offset", func() {
			path := tempDir + "/readat.txt"
			err := os.WriteFile(path, []byte("0123456789"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			buf := make([]byte, 3)
			n, err := p.ReadAt(buf, 5)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(3))
			Expect(string(buf)).To(Equal("567"))
		})

		It("should write at specific offset", func() {
			path := tempDir + "/writeat.txt"
			err := os.WriteFile(path, []byte("0123456789"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := New(path, os.O_RDWR, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			n, err := p.WriteAt([]byte("ABC"), 3)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(3))

			// Verify
			p.Seek(0, io.SeekStart)
			buf := make([]byte, 10)
			p.Read(buf)
			Expect(string(buf)).To(Equal("012ABC6789"))
		})
	})

	Describe("WriteString", func() {
		It("should write string successfully", func() {
			path := tempDir + "/writestring.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			n, err := p.WriteString("Hello, World!")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(13))

			// Verify
			p.Seek(0, io.SeekStart)
			buf := make([]byte, 13)
			p.Read(buf)
			Expect(string(buf)).To(Equal("Hello, World!"))
		})
	})

	Describe("WriteTo", func() {
		It("should write file contents to writer", func() {
			path := tempDir + "/writeto.txt"
			err := os.WriteFile(path, []byte("WriteTo test data"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Write to a buffer
			destPath := tempDir + "/writeto-dest.txt"
			dest, err := os.Create(destPath)
			Expect(err).ToNot(HaveOccurred())
			defer dest.Close()

			n, err := p.WriteTo(dest)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(17)))
		})
	})

	Describe("ReadFrom", func() {
		It("should read from source and write to file", func() {
			path := tempDir + "/readfrom.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			srcPath := tempDir + "/readfrom-src.txt"
			err = os.WriteFile(srcPath, []byte("ReadFrom source"), 0644)
			Expect(err).ToNot(HaveOccurred())

			src, err := os.Open(srcPath)
			Expect(err).ToNot(HaveOccurred())
			defer src.Close()

			n, err := p.ReadFrom(src)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(15)))

			// Verify
			p.Seek(0, io.SeekStart)
			buf := make([]byte, 15)
			p.Read(buf)
			Expect(string(buf)).To(Equal("ReadFrom source"))
		})
	})

	Describe("Reset manual trigger", func() {
		It("should manually trigger reset callback with specific size", func() {
			path := tempDir + "/reset-manual.txt"
			err := os.WriteFile(path, []byte("0123456789"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := New(path, os.O_RDWR, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			resetCalled := false
			var resetSize int64
			p.RegisterFctReset(func(max, current int64) {
				resetCalled = true
				resetSize = max
			})

			// Read to position
			buf := make([]byte, 5)
			p.Read(buf)

			// Manually trigger reset with specific size
			p.Reset(100)
			Expect(resetCalled).To(BeTrue())
			Expect(resetSize).To(Equal(int64(100)))
		})

		It("should auto-detect file size when max is 0", func() {
			path := tempDir + "/reset-auto.txt"
			err := os.WriteFile(path, []byte("0123456789"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := New(path, os.O_RDWR, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			resetCalled := false
			var resetSize int64
			p.RegisterFctReset(func(max, current int64) {
				resetCalled = true
				resetSize = max
			})

			// Read to position
			buf := make([]byte, 5)
			p.Read(buf)

			// Auto-detect size
			p.Reset(0)
			Expect(resetCalled).To(BeTrue())
			Expect(resetSize).To(Equal(int64(10)))
		})
	})
})
