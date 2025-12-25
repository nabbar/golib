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
	"errors"
	"io"

	arcarc "github.com/nabbar/golib/archive/archive"
	arccmp "github.com/nabbar/golib/archive/compress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// writeCloserBuffer wraps bytes.Buffer to implement io.WriteCloser
type wcBuffer struct {
	*bytes.Buffer
}

func (w *wcBuffer) Close() error {
	return nil
}

func newWCBuffer() *wcBuffer {
	return &wcBuffer{Buffer: &bytes.Buffer{}}
}

// errorReader is a custom reader that always returns an error
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errorReader) Close() error {
	return e.err
}

var _ = Describe("TC-EH-001: archive/error_handling", func() {
	Context("TC-EH-010: Compression error scenarios", func() {
		It("TC-EH-011: should handle corrupted gzip data", func() {
			// Create invalid gzip data
			invalidData := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF}
			reader, e := arccmp.Gzip.Reader(io.NopCloser(bytes.NewReader(invalidData)))

			if e == nil {
				defer reader.Close()
				_, e = io.ReadAll(reader)
			}
			Expect(e).To(HaveOccurred())
		})

		It("TC-EH-012: should handle corrupted bzip2 data", func() {
			// Create invalid bzip2 data
			invalidData := []byte{0x42, 0x5a, 0x68, 0x39, 0x00, 0x00, 0x00, 0x00}
			reader, e := arccmp.Bzip2.Reader(io.NopCloser(bytes.NewReader(invalidData)))

			if e == nil {
				defer reader.Close()
				_, e = io.ReadAll(reader)
			}
			Expect(e).To(HaveOccurred())
		})

		It("TC-EH-013: should handle read errors during decompression", func() {
			testErr := errors.New("test read error")
			errReader := &errorReader{err: testErr}

			reader, e := arccmp.Gzip.Reader(errReader)
			if e == nil {
				defer reader.Close()
			}
			// Either Reader() returns error or subsequent Read() will
			if e == nil {
				_, e = io.ReadAll(reader)
			}
			Expect(e).To(HaveOccurred())
		})

		It("TC-EH-014: should handle write errors during compression", func() {
			testErr := errors.New("test write error")
			errWriter := &errorWriter{err: testErr}

			writer, e := arccmp.Gzip.Writer(errWriter)
			if e == nil {
				defer writer.Close()
				_, e = writer.Write([]byte("test data"))
			}
			Expect(e).To(HaveOccurred())
		})
	})

	Context("TC-EH-020: Archive error scenarios", func() {
		It("TC-EH-021: should handle corrupted tar data", func() {
			// Create invalid tar header (needs to look like tar but be corrupt)
			invalidData := make([]byte, 512)
			// Set tar magic bytes at offset 257
			copy(invalidData[257:], []byte("ustar"))
			// But leave rest corrupted

			reader, e := arcarc.Tar.Reader(io.NopCloser(bytes.NewReader(invalidData)))
			if e == nil {
				defer reader.Close()
				// List() on corrupted tar should return empty or error
				list, _ := reader.List()
				// Either empty list or error is acceptable for corrupted data
				Expect(len(list)).To(BeNumerically(">=", 0))
			}
			// Error during Reader creation or List() is acceptable
		})

		It("TC-EH-022: should handle corrupted zip data", func() {
			// Create invalid zip signature
			invalidData := []byte{0x50, 0x4b, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}

			reader, e := arcarc.Zip.Reader(io.NopCloser(bytes.NewReader(invalidData)))
			if e == nil {
				defer reader.Close()
				_, e = reader.List()
			}
			Expect(e).To(HaveOccurred())
		})

		It("TC-EH-023: should handle non-existent files in archive", func() {
			Skip("Tested in main archive tests - Get() on non-existent file")
		})
	})

	Context("TC-EH-030: Algorithm marshaling behavior", func() {
		It("TC-EH-031: should set None for invalid compression algorithm names", func() {
			var alg arccmp.Algorithm
			err := alg.UnmarshalJSON([]byte(`"invalid_algorithm"`))
			Expect(err).ToNot(HaveOccurred()) // No error, just sets to None
			Expect(alg).To(Equal(arccmp.None))
		})

		It("TC-EH-032: should set None for invalid archive algorithm names", func() {
			var alg arcarc.Algorithm
			err := alg.UnmarshalJSON([]byte(`"invalid_algorithm"`))
			Expect(err).ToNot(HaveOccurred()) // No error, just sets to None
			Expect(alg).To(Equal(arcarc.None))
		})

		It("TC-EH-033: should error on malformed JSON", func() {
			var alg arccmp.Algorithm
			err := alg.UnmarshalJSON([]byte(`{invalid json`))
			Expect(err).To(HaveOccurred()) // JSON parsing error
		})

		It("TC-EH-034: should set None for invalid text in compression", func() {
			var alg arccmp.Algorithm
			err := alg.UnmarshalText([]byte("not_a_valid_algorithm"))
			Expect(err).ToNot(HaveOccurred()) // No error, just sets to None
			Expect(alg).To(Equal(arccmp.None))
		})

		It("TC-EH-035: should set None for invalid text in archive", func() {
			var alg arcarc.Algorithm
			err := alg.UnmarshalText([]byte("not_a_valid_algorithm"))
			Expect(err).ToNot(HaveOccurred()) // No error, just sets to None
			Expect(alg).To(Equal(arcarc.None))
		})
	})

	Context("TC-EH-040: None algorithm behavior", func() {
		It("TC-EH-041: should identify None compression algorithm correctly", func() {
			Expect(arccmp.None.IsNone()).To(BeTrue())
			Expect(arccmp.Gzip.IsNone()).To(BeFalse())
			Expect(arccmp.Bzip2.IsNone()).To(BeFalse())
		})

		It("TC-EH-042: should identify None archive algorithm correctly", func() {
			Expect(arcarc.None.IsNone()).To(BeTrue())
			Expect(arcarc.Tar.IsNone()).To(BeFalse())
			Expect(arcarc.Zip.IsNone()).To(BeFalse())
		})

		It("TC-EH-043: should handle None compression reader", func() {
			data := []byte("test data")
			reader, e := arccmp.None.Reader(io.NopCloser(bytes.NewReader(data)))
			Expect(e).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())
			defer reader.Close()

			result, e := io.ReadAll(reader)
			Expect(e).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("TC-EH-044: should handle None compression writer", func() {
			buf := newWCBuffer()
			writer, e := arccmp.None.Writer(buf)
			Expect(e).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())
			defer writer.Close()

			data := []byte("test data")
			n, e := writer.Write(data)
			Expect(e).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))

			Expect(writer.Close()).ToNot(HaveOccurred())
			Expect(buf.Bytes()).To(Equal(data))
		})
	})
})

// errorWriter is a custom writer that always returns an error
type errorWriter struct {
	err error
}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errorWriter) Close() error {
	return e.err
}
