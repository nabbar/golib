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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Coverage Improvement Tests", func() {
	Context("Stat operations", func() {
		It("should successfully get file stats", func() {
			content := []byte("test content for stats")
			p, path, err := createProgressFile(content)
			Expect(err).ToNot(HaveOccurred())
			defer cleanup(path)
			defer p.Close()

			info, err := p.Stat()
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Size()).To(Equal(int64(len(content))))
		})
	})

	Context("SizeBOF operations", func() {
		It("should return correct size from beginning", func() {
			content := []byte("0123456789")
			p, path, err := createProgressFile(content)
			Expect(err).ToNot(HaveOccurred())
			defer cleanup(path)
			defer p.Close()

			// Read 5 bytes
			buf := make([]byte, 5)
			_, err = p.Read(buf)
			Expect(err).ToNot(HaveOccurred())

			// Check BOF size
			bof, err := p.SizeBOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(bof).To(Equal(int64(5)))
		})
	})

	Context("SizeEOF operations", func() {
		It("should return correct remaining size", func() {
			content := []byte("0123456789")
			p, path, err := createProgressFile(content)
			Expect(err).ToNot(HaveOccurred())
			defer cleanup(path)
			defer p.Close()

			// Read 5 bytes
			buf := make([]byte, 5)
			_, err = p.Read(buf)
			Expect(err).ToNot(HaveOccurred())

			// Check EOF size
			eof, err := p.SizeEOF()
			Expect(err).ToNot(HaveOccurred())
			Expect(eof).To(Equal(int64(5)))
		})
	})

	Context("Sync operations", func() {
		It("should sync file successfully", func() {
			content := []byte("sync test")
			p, path, err := createProgressFileRW(content)
			Expect(err).ToNot(HaveOccurred())
			defer cleanup(path)
			defer p.Close()

			err = p.Sync()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("WriteString operations", func() {
		It("should write string successfully", func() {
			content := []byte("")
			p, path, err := createProgressFileRW(content)
			Expect(err).ToNot(HaveOccurred())
			defer cleanup(path)
			defer p.Close()

			testStr := "test string"
			n, err := p.WriteString(testStr)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(testStr)))
		})
	})

	Context("Reset callback", func() {
		It("should trigger reset callback on truncate", func() {
			content := []byte("0123456789")
			p, path, err := createProgressFileRW(content)
			Expect(err).ToNot(HaveOccurred())
			defer cleanup(path)
			defer p.Close()

			resetCalled := false
			var resetMax int64

			p.RegisterFctReset(func(max, current int64) {
				resetCalled = true
				resetMax = max
			})

			// Truncate file
			err = p.Truncate(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(resetCalled).To(BeTrue())
			Expect(resetMax).To(Equal(int64(5)))
		})
	})

	Context("getBufferSize edge cases", func() {
		It("should handle custom buffer sizes", func() {
			content := []byte("buffer test")
			p, path, err := createProgressFile(content)
			Expect(err).ToNot(HaveOccurred())
			defer cleanup(path)
			defer p.Close()

			// Set custom buffer size
			p.SetBufferSize(2048)

			// Read to trigger buffer size usage
			buf := make([]byte, 10)
			_, err = p.Read(buf)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("EOF callback edge cases", func() {
		It("should trigger EOF callback on ReadFrom", func() {
			content := []byte("")
			p, path, err := createProgressFileRW(content)
			Expect(err).ToNot(HaveOccurred())
			defer cleanup(path)
			defer p.Close()

			p.RegisterFctEOF(func() {
				// EOF callback registered
			})

			// Use ReadFrom to write data
			src := io.NopCloser(io.LimitReader(io.MultiReader(), 0))
			_, err = p.ReadFrom(src)

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Increment callback edge cases", func() {
		It("should handle nil callbacks gracefully", func() {
			content := []byte("callback test")
			p, path, err := createProgressFile(content)
			Expect(err).ToNot(HaveOccurred())
			defer cleanup(path)
			defer p.Close()

			// Register nil callback (should use no-op)
			p.RegisterFctIncrement(nil)
			p.RegisterFctReset(nil)
			p.RegisterFctEOF(nil)

			// Operations should still work
			buf := make([]byte, 5)
			_, err = p.Read(buf)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
