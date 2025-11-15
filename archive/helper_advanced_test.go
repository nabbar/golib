/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package archive_test

import (
	"bytes"
	"io"
	"strings"

	arccmp "github.com/nabbar/golib/archive/compress"
	archlp "github.com/nabbar/golib/archive/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// wcBufHelper wraps bytes.Buffer to implement io.WriteCloser
type wcBufHelper struct {
	*bytes.Buffer
}

func (w *wcBufHelper) Close() error {
	return nil
}

func newWCBufHelper() *wcBufHelper {
	return &wcBufHelper{Buffer: &bytes.Buffer{}}
}

var _ = Describe("archive/helper_advanced", func() {
	Context("Helper with different source types", func() {
		It("should create helper with Reader source", func() {
			src := bytes.NewReader([]byte("test data"))
			h, e := archlp.New(arccmp.Gzip, archlp.Compress, src)
			Expect(e).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("should create helper with Writer source", func() {
			var dst bytes.Buffer
			h, e := archlp.New(arccmp.Gzip, archlp.Compress, &dst)
			Expect(e).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("should return error for invalid source type", func() {
			_, e := archlp.New(arccmp.Gzip, archlp.Compress, "invalid source")
			Expect(e).To(Equal(archlp.ErrInvalidSource))
		})

		It("should return error for nil source", func() {
			_, e := archlp.New(arccmp.Gzip, archlp.Compress, nil)
			Expect(e).To(Equal(archlp.ErrInvalidSource))
		})
	})

	Context("NewReader specific tests", func() {
		It("should create compress reader", func() {
			src := bytes.NewReader([]byte("test data"))
			h, e := archlp.NewReader(arccmp.Gzip, archlp.Compress, src)
			Expect(e).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("should create decompress reader", func() {
			// First create compressed data
			buf := newWCBufHelper()
			w, _ := arccmp.Gzip.Writer(buf)
			w.Write([]byte("test data"))
			w.Close()

			// Now create decompress reader
			h, e := archlp.NewReader(arccmp.Gzip, archlp.Decompress, buf)
			Expect(e).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()

			// Read and verify
			result, e := io.ReadAll(h)
			Expect(e).ToNot(HaveOccurred())
			Expect(string(result)).To(Equal("test data"))
		})

		It("should return error for invalid operation", func() {
			src := bytes.NewReader([]byte("test"))
			_, e := archlp.NewReader(arccmp.Gzip, archlp.Operation(99), src)
			Expect(e).To(Equal(archlp.ErrInvalidOperation))
		})
	})

	Context("NewWriter specific tests", func() {
		It("should create compress writer", func() {
			var dst bytes.Buffer
			h, e := archlp.NewWriter(arccmp.Gzip, archlp.Compress, &dst)
			Expect(e).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()

			// Write data
			n, e := h.Write([]byte("test data"))
			Expect(e).ToNot(HaveOccurred())
			Expect(n).To(Equal(9))
		})

		It("should create decompress writer", func() {
			var dst bytes.Buffer
			h, e := archlp.NewWriter(arccmp.Gzip, archlp.Decompress, &dst)
			Expect(e).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("should return error for invalid operation", func() {
			var dst bytes.Buffer
			_, e := archlp.NewWriter(arccmp.Gzip, archlp.Operation(99), &dst)
			Expect(e).To(Equal(archlp.ErrInvalidOperation))
		})
	})

	Context("Complex data flow scenarios", func() {
		It("should handle large data streaming", func() {
			// Create large test data (100KB)
			largeData := strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 2000)

			for _, alg := range arccmp.List() {
				if alg.IsNone() {
					continue
				}

				src := bytes.NewReader([]byte(largeData))
				res := bytes.NewBuffer(make([]byte, 0))

				// Compress
				c, e := archlp.New(alg, archlp.Compress, src)
				Expect(e).NotTo(HaveOccurred())

				// Decompress
				d, e := archlp.New(alg, archlp.Decompress, c)
				Expect(e).NotTo(HaveOccurred())

				// Copy through the pipeline
				n, e := io.Copy(res, d)
				Expect(e).NotTo(HaveOccurred())
				Expect(n).To(BeNumerically("==", len(largeData)))

				Expect(d.Close()).NotTo(HaveOccurred())
				Expect(c.Close()).NotTo(HaveOccurred())

				// Verify data integrity
				Expect(res.String()).To(Equal(largeData))
			}
		})

		It("should handle incremental reads", func() {
			testData := "Hello, World! This is a test."

			for _, alg := range []arccmp.Algorithm{arccmp.Gzip, arccmp.Bzip2} {
				// Compress
				var compressed bytes.Buffer
				c, _ := archlp.New(alg, archlp.Compress, bytes.NewReader([]byte(testData)))
				io.Copy(&compressed, c)
				c.Close()

				// Decompress with small buffer reads
				d, e := archlp.NewReader(alg, archlp.Decompress, &compressed)
				Expect(e).ToNot(HaveOccurred())

				var result bytes.Buffer
				buf := make([]byte, 5) // Small buffer for incremental reads
				for {
					n, e := d.Read(buf)
					if n > 0 {
						result.Write(buf[:n])
					}
					if e == io.EOF {
						break
					}
					Expect(e).ToNot(HaveOccurred())
				}
				d.Close()

				Expect(result.String()).To(Equal(testData))
			}
		})

		It("should handle incremental writes", func() {
			testData := "Hello, World! This is a test."
			chunks := []string{"Hello, ", "World! ", "This is ", "a test."}

			for _, alg := range []arccmp.Algorithm{arccmp.Gzip, arccmp.Bzip2} {
				var compressed bytes.Buffer
				var decompressed bytes.Buffer

				// Create compression writer
				c, e := archlp.NewWriter(alg, archlp.Compress, &compressed)
				Expect(e).ToNot(HaveOccurred())

				// Write in chunks
				for _, chunk := range chunks {
					_, e = c.Write([]byte(chunk))
					Expect(e).ToNot(HaveOccurred())
				}
				Expect(c.Close()).ToNot(HaveOccurred())

				// Decompress
				d, e := archlp.NewReader(alg, archlp.Decompress, &compressed)
				Expect(e).ToNot(HaveOccurred())

				io.Copy(&decompressed, d)
				d.Close()

				Expect(decompressed.String()).To(Equal(testData))
			}
		})
	})

	Context("Edge cases", func() {
		It("should handle empty data compression", func() {
			for _, alg := range []arccmp.Algorithm{arccmp.Gzip, arccmp.Bzip2, arccmp.LZ4, arccmp.XZ} {
				src := bytes.NewReader([]byte{})
				res := bytes.NewBuffer(make([]byte, 0))

				c, e := archlp.New(alg, archlp.Compress, src)
				Expect(e).NotTo(HaveOccurred())

				d, e := archlp.New(alg, archlp.Decompress, c)
				Expect(e).NotTo(HaveOccurred())

				n, e := io.Copy(res, d)
				Expect(e).NotTo(HaveOccurred())
				Expect(n).To(BeEquivalentTo(0))

				d.Close()
				c.Close()

				Expect(res.Len()).To(BeEquivalentTo(0))
			}
		})

		It("should handle single byte data", func() {
			for _, alg := range []arccmp.Algorithm{arccmp.Gzip, arccmp.Bzip2} {
				src := bytes.NewReader([]byte{'A'})
				res := bytes.NewBuffer(make([]byte, 0))

				c, e := archlp.New(alg, archlp.Compress, src)
				Expect(e).NotTo(HaveOccurred())

				d, e := archlp.New(alg, archlp.Decompress, c)
				Expect(e).NotTo(HaveOccurred())

				io.Copy(res, d)
				d.Close()
				c.Close()

				Expect(res.Bytes()).To(Equal([]byte{'A'}))
			}
		})
	})

	Context("Operation type", func() {
		It("should have correct operation type values", func() {
			Expect(int(archlp.Compress)).To(BeNumerically(">=", 0))
			Expect(int(archlp.Decompress)).To(BeNumerically(">=", 0))
			Expect(archlp.Compress).ToNot(Equal(archlp.Decompress))
		})
	})
})
