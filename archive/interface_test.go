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

	libarc "github.com/nabbar/golib/archive"
	arcarc "github.com/nabbar/golib/archive/archive"
	arccmp "github.com/nabbar/golib/archive/compress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// wcBuf wraps bytes.Buffer to implement io.WriteCloser
type wcBuf struct {
	*bytes.Buffer
}

func (w *wcBuf) Close() error {
	return nil
}

func newWCBuf() *wcBuf {
	return &wcBuf{Buffer: &bytes.Buffer{}}
}

var _ = Describe("TC-IF-001: archive/interface", func() {
	Context("TC-IF-010: ParseCompression function", func() {
		It("TC-IF-011: should parse valid compression algorithm names", func() {
			Expect(libarc.ParseCompression("gzip")).To(Equal(arccmp.Gzip))
			Expect(libarc.ParseCompression("bzip2")).To(Equal(arccmp.Bzip2))
			Expect(libarc.ParseCompression("lz4")).To(Equal(arccmp.LZ4))
			Expect(libarc.ParseCompression("xz")).To(Equal(arccmp.XZ))
		})

		It("TC-IF-012: should parse case-insensitive algorithm names", func() {
			Expect(libarc.ParseCompression("GZIP")).To(Equal(arccmp.Gzip))
			Expect(libarc.ParseCompression("GZip")).To(Equal(arccmp.Gzip))
			Expect(libarc.ParseCompression("BZip2")).To(Equal(arccmp.Bzip2))
		})

		It("TC-IF-013: should return None for invalid algorithm names", func() {
			Expect(libarc.ParseCompression("invalid")).To(Equal(arccmp.None))
			Expect(libarc.ParseCompression("")).To(Equal(arccmp.None))
			Expect(libarc.ParseCompression("zip")).To(Equal(arccmp.None)) // zip is archive, not compression
		})
	})

	Context("TC-IF-020: ParseArchive function", func() {
		It("TC-IF-021: should parse valid archive algorithm names", func() {
			Expect(libarc.ParseArchive("tar")).To(Equal(arcarc.Tar))
			Expect(libarc.ParseArchive("zip")).To(Equal(arcarc.Zip))
		})

		It("TC-IF-022: should parse case-insensitive algorithm names", func() {
			Expect(libarc.ParseArchive("TAR")).To(Equal(arcarc.Tar))
			Expect(libarc.ParseArchive("ZIP")).To(Equal(arcarc.Zip))
			Expect(libarc.ParseArchive("Tar")).To(Equal(arcarc.Tar))
		})

		It("TC-IF-023: should return None for invalid algorithm names", func() {
			Expect(libarc.ParseArchive("invalid")).To(Equal(arcarc.None))
			Expect(libarc.ParseArchive("")).To(Equal(arcarc.None))
			Expect(libarc.ParseArchive("gzip")).To(Equal(arcarc.None)) // gzip is compression, not archive
		})
	})

	Context("TC-IF-030: DetectCompression function", func() {
		It("TC-IF-031: should detect gzip compression", func() {
			// Create a simple gzip compressed data
			buf := newWCBuf()
			writer, e := arccmp.Gzip.Writer(buf)
			Expect(e).ToNot(HaveOccurred())

			_, e = writer.Write([]byte("test data"))
			Expect(e).ToNot(HaveOccurred())
			Expect(writer.Close()).ToNot(HaveOccurred())

			// Detect the compression
			alg, reader, e := libarc.DetectCompression(bytes.NewReader(buf.Bytes()))
			Expect(e).ToNot(HaveOccurred())
			Expect(alg).To(Equal(arccmp.Gzip))
			Expect(reader).ToNot(BeNil())
			defer reader.Close()

			// Verify we can read the decompressed data
			decompressed, e := io.ReadAll(reader)
			Expect(e).ToNot(HaveOccurred())
			Expect(string(decompressed)).To(Equal("test data"))
		})

		It("TC-IF-032: should detect bzip2 compression", func() {
			// Create a simple bzip2 compressed data
			buf := newWCBuf()
			writer, e := arccmp.Bzip2.Writer(buf)
			Expect(e).ToNot(HaveOccurred())

			_, e = writer.Write([]byte("test data"))
			Expect(e).ToNot(HaveOccurred())
			Expect(writer.Close()).ToNot(HaveOccurred())

			// Detect the compression
			alg, reader, e := libarc.DetectCompression(bytes.NewReader(buf.Bytes()))
			Expect(e).ToNot(HaveOccurred())
			Expect(alg).To(Equal(arccmp.Bzip2))
			Expect(reader).ToNot(BeNil())
			defer reader.Close()

			// Verify we can read the decompressed data
			decompressed, e := io.ReadAll(reader)
			Expect(e).ToNot(HaveOccurred())
			Expect(string(decompressed)).To(Equal("test data"))
		})

		It("TC-IF-033: should return None for uncompressed data", func() {
			data := []byte("plain text data")
			alg, reader, e := libarc.DetectCompression(bytes.NewReader(data))
			Expect(e).ToNot(HaveOccurred())
			Expect(alg).To(Equal(arccmp.None))
			Expect(reader).ToNot(BeNil())
			defer reader.Close()
		})

		It("TC-IF-034: should handle empty input gracefully", func() {
			_, _, e := libarc.DetectCompression(bytes.NewReader([]byte{}))
			Expect(e).To(HaveOccurred()) // EOF error expected
		})
	})

	Context("TC-IF-040: DetectArchive function", func() {
		It("TC-IF-041: should return None for unarchived data", func() {
			// Need enough data for header detection (at least 265 bytes)
			data := bytes.Repeat([]byte("plain text data that is not an archive "), 10)
			alg, reader, closer, e := libarc.DetectArchive(io.NopCloser(bytes.NewReader(data)))

			Expect(e).ToNot(HaveOccurred())
			Expect(alg).To(Equal(arcarc.None))
			Expect(reader).To(BeNil())
			Expect(closer).ToNot(BeNil())
			defer closer.Close()
		})

		It("TC-IF-042: should handle small input", func() {
			// Small input (less than header size) should error
			_, _, _, e := libarc.DetectArchive(io.NopCloser(bytes.NewReader([]byte("small"))))
			Expect(e).To(HaveOccurred()) // EOF error expected when trying to peek
		})
	})
})
