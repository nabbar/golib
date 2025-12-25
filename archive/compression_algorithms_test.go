/*
 *  MIT License
 *
 *  Copyright (c) 2025 Nicolas JUHEL
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

	arccmp "github.com/nabbar/golib/archive/compress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// writeCloserBuffer wraps bytes.Buffer to implement io.WriteCloser
type writeCloserBuffer struct {
	*bytes.Buffer
}

func (w *writeCloserBuffer) Close() error {
	return nil
}

func newWriteCloserBuffer() *writeCloserBuffer {
	return &writeCloserBuffer{Buffer: &bytes.Buffer{}}
}

var _ = Describe("TC-CA-001: archive/compression_algorithms", func() {
	// Test data of various sizes
	testCases := []struct {
		name string
		data string
	}{
		{"empty", ""},
		{"small", "Hello, World!"},
		{"medium", string(make([]byte, 1024))}, // 1KB of zeros
		{"large", loremIpsum[:10000]},          // ~10KB of text
		{"repeated", string(bytes.Repeat([]byte("A"), 5000))},
		{"binary", string([]byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC})},
	}

	for _, tc := range testCases {
		Context("TC-CA-010: Testing with "+tc.name+" data", func() {
			It("TC-CA-011: should compress and decompress with Gzip", func() {
				testCompressionAlgorithm(arccmp.Gzip, tc.data)
			})

			It("TC-CA-012: should compress and decompress with Bzip2", func() {
				testCompressionAlgorithm(arccmp.Bzip2, tc.data)
			})

			It("TC-CA-013: should compress and decompress with LZ4", func() {
				testCompressionAlgorithm(arccmp.LZ4, tc.data)
			})

			It("TC-CA-014: should compress and decompress with XZ", func() {
				testCompressionAlgorithm(arccmp.XZ, tc.data)
			})
		})
	}

	Context("TC-CA-020: Compression efficiency", func() {
		It("TC-CA-021: should compress repeated data efficiently", func() {
			// Repeated data should compress well
			data := bytes.Repeat([]byte("test"), 1000)

			for _, alg := range []arccmp.Algorithm{arccmp.Gzip, arccmp.Bzip2, arccmp.LZ4, arccmp.XZ} {
				compressed := newWriteCloserBuffer()
				writer, e := alg.Writer(compressed)
				Expect(e).ToNot(HaveOccurred())

				_, e = writer.Write(data)
				Expect(e).ToNot(HaveOccurred())
				Expect(writer.Close()).ToNot(HaveOccurred())

				// Compressed size should be significantly smaller than original
				// (at least 50% compression for such highly repetitive data)
				Expect(compressed.Len()).To(BeNumerically("<", len(data)/2),
					"Algorithm %s should compress repetitive data efficiently", alg.String())
			}
		})

		It("TC-CA-022: should handle incompressible data gracefully", func() {
			// Random-like data (lorem ipsum) compresses poorly
			data := []byte(loremIpsum[:1000])

			for _, alg := range []arccmp.Algorithm{arccmp.Gzip, arccmp.Bzip2, arccmp.LZ4, arccmp.XZ} {
				compressed := newWriteCloserBuffer()
				writer, e := alg.Writer(compressed)
				Expect(e).ToNot(HaveOccurred())

				_, e = writer.Write(data)
				Expect(e).ToNot(HaveOccurred())
				Expect(writer.Close()).ToNot(HaveOccurred())

				// Compressed data might be larger due to overhead, but should still work
				Expect(compressed.Len()).To(BeNumerically(">", 0))
			}
		})
	})

	Context("TC-CA-030: Multiple write operations", func() {
		It("TC-CA-031: should handle multiple writes to compressor", func() {
			for _, alg := range []arccmp.Algorithm{arccmp.Gzip, arccmp.Bzip2, arccmp.LZ4, arccmp.XZ} {
				compressed := newWriteCloserBuffer()
				writer, e := alg.Writer(compressed)
				Expect(e).ToNot(HaveOccurred())

				// Write data in chunks
				chunks := []string{"Hello", ", ", "World", "!"}
				for _, chunk := range chunks {
					_, e = writer.Write([]byte(chunk))
					Expect(e).ToNot(HaveOccurred())
				}
				Expect(writer.Close()).ToNot(HaveOccurred())

				// Decompress and verify
				reader, e := alg.Reader(io.NopCloser(compressed))
				Expect(e).ToNot(HaveOccurred())
				defer reader.Close()

				result, e := io.ReadAll(reader)
				Expect(e).ToNot(HaveOccurred())
				Expect(string(result)).To(Equal("Hello, World!"))
			}
		})
	})

	Context("TC-CA-040: Algorithm properties", func() {
		It("TC-CA-041: should return correct extensions", func() {
			Expect(arccmp.Gzip.Extension()).To(Equal(".gz"))
			Expect(arccmp.Bzip2.Extension()).To(Equal(".bz2"))
			Expect(arccmp.LZ4.Extension()).To(Equal(".lz4"))
			Expect(arccmp.XZ.Extension()).To(Equal(".xz"))
			Expect(arccmp.None.Extension()).To(Equal(""))
		})

		It("TC-CA-042: should return correct string representation", func() {
			Expect(arccmp.Gzip.String()).To(Equal("gzip"))
			Expect(arccmp.Bzip2.String()).To(Equal("bzip2"))
			Expect(arccmp.LZ4.String()).To(Equal("lz4"))
			Expect(arccmp.XZ.String()).To(Equal("xz"))
			Expect(arccmp.None.String()).To(Equal("none"))
		})

		It("TC-CA-043: should list all available algorithms", func() {
			algorithms := arccmp.List()
			Expect(algorithms).To(ContainElement(arccmp.Gzip))
			Expect(algorithms).To(ContainElement(arccmp.Bzip2))
			Expect(algorithms).To(ContainElement(arccmp.LZ4))
			Expect(algorithms).To(ContainElement(arccmp.XZ))
			Expect(algorithms).To(ContainElement(arccmp.None))
		})
	})

	Context("TC-CA-050: Header detection", func() {
		It("TC-CA-051: should detect Gzip header correctly", func() {
			buf := newWriteCloserBuffer()
			writer, _ := arccmp.Gzip.Writer(buf)
			writer.Write([]byte("test"))
			writer.Close()

			header := buf.Bytes()[:6]
			Expect(arccmp.Gzip.DetectHeader(header)).To(BeTrue())
			Expect(arccmp.Bzip2.DetectHeader(header)).To(BeFalse())
		})

		It("TC-CA-052: should detect Bzip2 header correctly", func() {
			buf := newWriteCloserBuffer()
			writer, _ := arccmp.Bzip2.Writer(buf)
			writer.Write([]byte("test"))
			writer.Close()

			header := buf.Bytes()[:6]
			Expect(arccmp.Bzip2.DetectHeader(header)).To(BeTrue())
			Expect(arccmp.Gzip.DetectHeader(header)).To(BeFalse())
		})

		It("TC-CA-053: should detect LZ4 header correctly", func() {
			buf := newWriteCloserBuffer()
			writer, _ := arccmp.LZ4.Writer(buf)
			writer.Write([]byte("test"))
			writer.Close()

			header := buf.Bytes()[:6]
			Expect(arccmp.LZ4.DetectHeader(header)).To(BeTrue())
			Expect(arccmp.Gzip.DetectHeader(header)).To(BeFalse())
		})

		It("TC-CA-054: should detect XZ header correctly", func() {
			buf := newWriteCloserBuffer()
			writer, _ := arccmp.XZ.Writer(buf)
			writer.Write([]byte("test"))
			writer.Close()

			header := buf.Bytes()[:6]
			Expect(arccmp.XZ.DetectHeader(header)).To(BeTrue())
			Expect(arccmp.Gzip.DetectHeader(header)).To(BeFalse())
		})
	})
})

// testCompressionAlgorithm is a helper function that tests compression/decompression roundtrip
func testCompressionAlgorithm(alg arccmp.Algorithm, data string) {
	compressed := newWriteCloserBuffer()

	// Compress
	writer, e := alg.Writer(compressed)
	Expect(e).ToNot(HaveOccurred())
	Expect(writer).ToNot(BeNil())

	n, e := writer.Write([]byte(data))
	Expect(e).ToNot(HaveOccurred())
	Expect(n).To(Equal(len(data)))

	e = writer.Close()
	Expect(e).ToNot(HaveOccurred())

	// Decompress
	reader, e := alg.Reader(io.NopCloser(compressed))
	Expect(e).ToNot(HaveOccurred())
	Expect(reader).ToNot(BeNil())
	defer reader.Close()

	decompressed, e := io.ReadAll(reader)
	Expect(e).ToNot(HaveOccurred())
	Expect(string(decompressed)).To(Equal(data))
}
