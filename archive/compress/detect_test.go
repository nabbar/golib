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

var _ = Describe("TC-DT-001: Detection Functions", func() {
	Context("TC-DT-002: DetectOnly function", func() {
		It("TC-DT-003: should detect Gzip format", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Gzip, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			alg, reader, err := compress.DetectOnly(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
			Expect(reader).ToNot(BeNil())
			defer reader.Close()

			// Verify peeked data is preserved
			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal(compressed))
		})

		It("TC-DT-004: should detect Bzip2 format", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Bzip2, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			alg, reader, err := compress.DetectOnly(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Bzip2))
			Expect(reader).ToNot(BeNil())
			defer reader.Close()
		})

		It("TC-DT-005: should detect LZ4 format", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.LZ4, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			alg, reader, err := compress.DetectOnly(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.LZ4))
			Expect(reader).ToNot(BeNil())
			defer reader.Close()
		})

		It("TC-DT-006: should detect XZ format", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.XZ, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			alg, reader, err := compress.DetectOnly(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.XZ))
			Expect(reader).ToNot(BeNil())
			defer reader.Close()
		})

		It("TC-DT-007: should return None for uncompressed data", func() {
			testData := newTestData(100)
			alg, reader, err := compress.DetectOnly(bytes.NewReader(testData.dat))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.None))
			Expect(reader).ToNot(BeNil())
			defer reader.Close()
		})

		It("TC-DT-008: should return error for insufficient data", func() {
			shortData := []byte{0x1F, 0x8B}
			_, _, err := compress.DetectOnly(bytes.NewReader(shortData))
			Expect(err).To(HaveOccurred())
		})

		It("TC-DT-009: should return error for empty reader", func() {
			_, _, err := compress.DetectOnly(bytes.NewReader([]byte{}))
			Expect(err).To(HaveOccurred())
		})

		It("TC-DT-010: should preserve all data after peek", func() {
			testData := newTestData(100)
			original := testData.dat

			alg, reader, err := compress.DetectOnly(bytes.NewReader(original))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.None))
			defer reader.Close()

			// Read all data and verify it matches original
			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal(original))
		})
	})

	Context("TC-DT-011: Detect function", func() {
		It("TC-DT-012: should detect and decompress Gzip", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Gzip, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			alg, reader, err := compress.Detect(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
			defer reader.Close()

			// Read decompressed data
			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-DT-013: should detect and decompress Bzip2", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Bzip2, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			alg, reader, err := compress.Detect(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Bzip2))
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-DT-014: should detect and decompress LZ4", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.LZ4, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			alg, reader, err := compress.Detect(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.LZ4))
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-DT-015: should detect and decompress XZ", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.XZ, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			alg, reader, err := compress.Detect(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.XZ))
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})

		It("TC-DT-016: should pass through None format", func() {
			testData := newTestData(100)
			alg, reader, err := compress.Detect(bytes.NewReader(testData.dat))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.None))
			defer reader.Close()

			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal(testData.dat))
		})

		It("TC-DT-017: should return error for insufficient data", func() {
			shortData := []byte{0x1F}
			_, _, err := compress.Detect(bytes.NewReader(shortData))
			Expect(err).To(HaveOccurred())
		})

		It("TC-DT-018: should handle large compressed data", func() {
			testData := newTestData(10000)
			compressed, err := compressTestData(compress.Gzip, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			alg, reader, err := compress.Detect(bytes.NewReader(compressed))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(testData.dat))
		})
	})

	Context("TC-DT-019: Edge cases", func() {
		It("TC-DT-020: should handle data starting with similar bytes", func() {
			// Data that starts with 0x1F but not gzip
			fakeGzip := []byte{0x1F, 0x00, 0x00, 0x00, 0x00, 0x00}
			alg, reader, err := compress.DetectOnly(bytes.NewReader(fakeGzip))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.None))
			defer reader.Close()
		})

		It("TC-DT-021: should handle exactly 6 bytes", func() {
			header := []byte{0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00}
			alg, reader, err := compress.DetectOnly(bytes.NewReader(header))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
			defer reader.Close()
		})

		It("TC-DT-022: should detect with extra data after header", func() {
			testData := newTestData(100)
			compressed, err := compressTestData(compress.Gzip, testData.dat)
			Expect(err).ToNot(HaveOccurred())

			// Add extra data
			withExtra := append(compressed, []byte("extra data")...)

			alg, _, err := compress.DetectOnly(bytes.NewReader(withExtra))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
		})
	})
})
