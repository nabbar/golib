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

package compress_test

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/compress"
)

var _ = Describe("TC-IO-001: I/O Operations", func() {
	Context("TC-IO-002: Reader method", func() {
		It("TC-IO-003: should create Gzip reader", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Gzip, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			reader, err := compress.Gzip.Reader(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-IO-004: should create Bzip2 reader", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Bzip2, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			reader, err := compress.Bzip2.Reader(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-IO-005: should create LZ4 reader", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.LZ4, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			reader, err := compress.LZ4.Reader(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-IO-006: should create XZ reader", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.XZ, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			reader, err := compress.XZ.Reader(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-IO-007: should create None reader (pass-through)", func() {
			testData := newTestData(100)
			reader, err := compress.None.Reader(bytes.NewReader(testData.dat))
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())
			defer reader.Close()

			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal(testData.dat))
		})

		It("TC-IO-008: should return error for invalid Gzip data", func() {
			invalidData := []byte{0x1F, 0x8B, 0xFF, 0xFF, 0xFF, 0xFF}
			reader, err := compress.Gzip.Reader(bytes.NewReader(invalidData))
			if err == nil {
				defer reader.Close()
				_, err = io.ReadAll(reader)
			}
			Expect(err).To(HaveOccurred())
		})

		It("TC-IO-009: should handle large data with Gzip", func() {
			testData := newTestData(10000)
			compressed, err := compressTestData(compress.Gzip, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			reader, err := compress.Gzip.Reader(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-IO-010: should handle empty data", func() {
			emptyData := []byte{}
			reader, err := compress.None.Reader(bytes.NewReader(emptyData))
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(BeEmpty())
		})
	})

	Context("TC-IO-011: Writer method", func() {
		It("TC-IO-012: should create Gzip writer", func() {
			testData := newTestData(100)
			var buf bytes.Buffer

			writer, err := compress.Gzip.Writer(nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())

			n, err := writer.Write(testData.dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(testData.dat)))

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())

			// Verify compressed data can be decompressed
			reader, err := compress.Gzip.Reader(&buf)
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-IO-013: should create Bzip2 writer", func() {
			testData := newTestData(100)
			var buf bytes.Buffer

			writer, err := compress.Bzip2.Writer(nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())

			n, err := writer.Write(testData.dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(testData.dat)))

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())

			// Verify compressed data
			reader, err := compress.Bzip2.Reader(&buf)
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-IO-014: should create LZ4 writer", func() {
			testData := newTestData(100)
			var buf bytes.Buffer

			writer, err := compress.LZ4.Writer(nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())

			n, err := writer.Write(testData.dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(testData.dat)))

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())

			// Verify compressed data
			reader, err := compress.LZ4.Reader(&buf)
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-IO-015: should create XZ writer", func() {
			testData := newTestData(100)
			var buf bytes.Buffer

			writer, err := compress.XZ.Writer(nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())

			n, err := writer.Write(testData.dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(testData.dat)))

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())

			// Verify compressed data
			reader, err := compress.XZ.Reader(&buf)
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-IO-016: should create None writer (pass-through)", func() {
			testData := newTestData(100)
			var buf bytes.Buffer

			writer, err := compress.None.Writer(nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())

			n, err := writer.Write(testData.dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(testData.dat)))

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.Bytes()).To(Equal(testData.dat))
		})

		It("TC-IO-017: should handle multiple writes", func() {
			testData1 := newTestData(50)
			testData2 := newTestData(50)
			var buf bytes.Buffer

			writer, err := compress.Gzip.Writer(nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())
			defer writer.Close()

			_, err = writer.Write(testData1.dat)
			Expect(err).ToNot(HaveOccurred())

			_, err = writer.Write(testData2.dat)
			Expect(err).ToNot(HaveOccurred())

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())

			// Verify
			reader, err := compress.Gzip.Reader(&buf)
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())

			expected := append(testData1.dat, testData2.dat...)
			Expect(decompressed).To(Equal(expected))
		})

		It("TC-IO-018: should handle large data with writer", func() {
			testData := newTestData(10000)
			var buf bytes.Buffer

			writer, err := compress.Gzip.Writer(nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())

			_, err = writer.Write(testData.dat)
			Expect(err).ToNot(HaveOccurred())

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())

			// Verify
			reader, err := compress.Gzip.Reader(&buf)
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-IO-019: should handle empty write", func() {
			emptyData := []byte{}
			var buf bytes.Buffer

			writer, err := compress.Gzip.Writer(nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())

			n, err := writer.Write(emptyData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("TC-IO-020: Round-trip tests", func() {
		It("TC-IO-021: should round-trip Gzip compression", func() {
			testData := newTestData(500)
			result, err := roundTripTest(compress.Gzip.Writer, compress.Gzip.Reader, testData.dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(testData.dat))
		})

		It("TC-IO-022: should round-trip Bzip2 compression", func() {
			testData := newTestData(500)
			result, err := roundTripTest(compress.Bzip2.Writer, compress.Bzip2.Reader, testData.dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(testData.dat))
		})

		It("TC-IO-023: should round-trip LZ4 compression", func() {
			testData := newTestData(500)
			result, err := roundTripTest(compress.LZ4.Writer, compress.LZ4.Reader, testData.dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(testData.dat))
		})

		It("TC-IO-024: should round-trip XZ compression", func() {
			testData := newTestData(500)
			result, err := roundTripTest(compress.XZ.Writer, compress.XZ.Reader, testData.dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(testData.dat))
		})

		It("TC-IO-025: should round-trip None (no compression)", func() {
			testData := newTestData(500)
			result, err := roundTripTest(compress.None.Writer, compress.None.Reader, testData.dat)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(testData.dat))
		})

		It("TC-IO-026: should round-trip all algorithms", func() {
			testData := newTestData(200)
			for _, alg := range compress.List() {
				if alg == compress.None {
					continue
				}
				result, err := roundTripTest(alg.Writer, alg.Reader, testData.dat)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(testData.dat), "Algorithm: %s", alg.String())
			}
		})

		It("TC-IO-027: should round-trip with various data sizes", func() {
			sizes := []int{0, 1, 10, 100, 1000, 5000}
			for _, size := range sizes {
				testData := newTestData(size)
				result, err := roundTripTest(compress.Gzip.Writer, compress.Gzip.Reader, testData.dat)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(testData.dat), "Size: %d", size)
			}
		})
	})

	Context("TC-IO-028: Edge cases", func() {
		It("TC-IO-029: should handle writer close without write", func() {
			var buf bytes.Buffer
			writer, err := compress.Gzip.Writer(nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())

			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-IO-030: should handle reader on empty compressed data", func() {
			var buf bytes.Buffer
			writer, err := compress.Gzip.Writer(nopWriteCloser{&buf})
			Expect(err).ToNot(HaveOccurred())
			writer.Close()

			reader, err := compress.Gzip.Reader(&buf)
			Expect(err).ToNot(HaveOccurred())
			defer reader.Close()

			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(BeEmpty())
		})

		It("TC-IO-031: should handle binary data", func() {
			binaryData := make([]byte, 256)
			for i := range binaryData {
				binaryData[i] = byte(i)
			}

			result, err := roundTripTest(compress.Gzip.Writer, compress.Gzip.Reader, binaryData)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(binaryData))
		})

		It("TC-IO-032: should handle highly compressible data", func() {
			// All zeros - very compressible
			compressible := bytes.Repeat([]byte{0}, 1000)

			result, err := roundTripTest(compress.Gzip.Writer, compress.Gzip.Reader, compressible)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(compressible))
		})

		It("TC-IO-033: should handle incompressible data", func() {
			// Random-like data - incompressible
			incompressible := make([]byte, 1000)
			for i := range incompressible {
				incompressible[i] = byte(i * 7 % 256)
			}

			result, err := roundTripTest(compress.Gzip.Writer, compress.Gzip.Reader, incompressible)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(incompressible))
		})
	})
})
