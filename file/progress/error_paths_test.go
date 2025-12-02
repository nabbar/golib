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

var _ = Describe("Error Paths Coverage", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "progress-error-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("Nil checks in I/O operations", func() {
		It("should return error when Read on closed file", func() {
			path := tempDir + "/closed.txt"
			err := os.WriteFile(path, []byte("data"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			// Try to read from closed file
			buf := make([]byte, 10)
			_, err = p.Read(buf)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Write on closed file", func() {
			path := tempDir + "/closed-write.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			// Try to write to closed file
			_, err = p.Write([]byte("data"))
			Expect(err).To(HaveOccurred())
		})

		It("should return error when ReadAt on closed file", func() {
			path := tempDir + "/closed-readat.txt"
			err := os.WriteFile(path, []byte("data"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			buf := make([]byte, 4)
			_, err = p.ReadAt(buf, 0)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when WriteAt on closed file", func() {
			path := tempDir + "/closed-writeat.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			_, err = p.WriteAt([]byte("data"), 0)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when WriteString on closed file", func() {
			path := tempDir + "/closed-writestring.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			_, err = p.WriteString("data")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when ReadFrom on closed file", func() {
			path := tempDir + "/closed-readfrom.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			src := io.NopCloser(io.LimitReader(io.MultiReader(), 0))
			_, err = p.ReadFrom(src)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when WriteTo on closed file", func() {
			path := tempDir + "/closed-writeto.txt"
			err := os.WriteFile(path, []byte("data"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			dest, _ := os.Create(tempDir + "/dest.txt")
			defer dest.Close()

			_, err = p.WriteTo(dest)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Seek on closed file", func() {
			path := tempDir + "/closed-seek.txt"
			err := os.WriteFile(path, []byte("data"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			_, err = p.Seek(0, io.SeekStart)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Stat on closed file", func() {
			path := tempDir + "/closed-stat.txt"
			err := os.WriteFile(path, []byte("data"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			_, err = p.Stat()
			Expect(err).To(HaveOccurred())
		})

		It("should return error when SizeBOF on closed file", func() {
			path := tempDir + "/closed-sizebof.txt"
			err := os.WriteFile(path, []byte("data"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			_, err = p.SizeBOF()
			Expect(err).To(HaveOccurred())
		})

		It("should return error when SizeEOF on closed file", func() {
			path := tempDir + "/closed-sizeeof.txt"
			err := os.WriteFile(path, []byte("data"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			_, err = p.SizeEOF()
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Truncate on closed file", func() {
			path := tempDir + "/closed-truncate.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			err = p.Truncate(0)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Sync on closed file", func() {
			path := tempDir + "/closed-sync.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			p.Close()

			err = p.Sync()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ReadFrom with nil reader", func() {
		It("should return error with nil reader", func() {
			path := tempDir + "/readfrom-nil.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			_, err = p.ReadFrom(nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("WriteTo with nil writer", func() {
		It("should return error with nil writer", func() {
			path := tempDir + "/writeto-nil.txt"
			err := os.WriteFile(path, []byte("data"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			_, err = p.WriteTo(nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ReadByte edge cases", func() {
		It("should handle read errors correctly", func() {
			path := tempDir + "/readbyte-edge.txt"
			err := os.WriteFile(path, []byte("X"), 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Read the byte
			b, err := p.ReadByte()
			Expect(err).ToNot(HaveOccurred())
			Expect(b).To(Equal(byte('X')))

			// Try to read past EOF
			_, err = p.ReadByte()
			Expect(err).To(Equal(io.EOF))
		})
	})

	Describe("WriteByte edge cases", func() {
		It("should handle seek operations during write", func() {
			path := tempDir + "/writebyte-edge.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Write first byte
			err = p.WriteByte('A')
			Expect(err).ToNot(HaveOccurred())

			// Write second byte
			err = p.WriteByte('B')
			Expect(err).ToNot(HaveOccurred())

			// Verify
			p.Seek(0, io.SeekStart)
			buf := make([]byte, 2)
			p.Read(buf)
			Expect(string(buf)).To(Equal("AB"))
		})
	})
})
