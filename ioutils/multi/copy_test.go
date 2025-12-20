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

package multi_test

import (
	"bytes"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/ioutils/multi"
)

// Tests for Multi copy operations and integration scenarios.
// These tests verify the Copy() method and demonstrate integration
// of read/write/copy operations in realistic workflows.
var _ = Describe("[TC-CP] Multi Copy Operations", func() {
	var m multi.Multi

	BeforeEach(func() {
		m = multi.New(false, false, multi.DefaultConfig())
	})

	Describe("Copy", func() {
		Context("copying to single writer", func() {
			It("[TC-CP-001] should copy data from reader to writer", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				input := io.NopCloser(strings.NewReader("test data"))
				m.SetInput(input)

				n, err := m.Copy()
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(int64(9)))
				Expect(buf.String()).To(Equal("test data"))
			})

			It("[TC-CP-001] should handle empty input", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				input := io.NopCloser(strings.NewReader(""))
				m.SetInput(input)

				n, err := m.Copy()
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(int64(0)))
				Expect(buf.String()).To(BeEmpty())
			})
		})

		Context("copying to multiple writers", func() {
			It("[TC-CP-002] should copy to all writers simultaneously", func() {
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)

				input := io.NopCloser(strings.NewReader("broadcast"))
				m.SetInput(input)

				n, err := m.Copy()
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(int64(9)))
				Expect(buf1.String()).To(Equal("broadcast"))
				Expect(buf2.String()).To(Equal("broadcast"))
				Expect(buf3.String()).To(Equal("broadcast"))
			})
		})

		Context("copying large data", func() {
			It("[TC-CP-003] should handle large data copy", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				largeData := strings.Repeat("x", 1024*1024) // 1MB
				input := io.NopCloser(strings.NewReader(largeData))
				m.SetInput(input)

				n, err := m.Copy()
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(int64(1024 * 1024)))
				Expect(buf.Len()).To(Equal(1024 * 1024))
			})
		})

		Context("copying with errors", func() {
			It("[TC-CP-004] should return error from reader", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				errorReader := &errorReader{err: io.ErrUnexpectedEOF}
				m.SetInput(io.NopCloser(errorReader))

				n, err := m.Copy()
				Expect(err).To(Equal(io.ErrUnexpectedEOF))
				Expect(n).To(Equal(int64(0)))
			})

			It("[TC-CP-004] should return error from writer", func() {
				errorWriter := &errorWriter{err: io.ErrShortWrite}
				m.AddWriter(errorWriter)

				input := io.NopCloser(strings.NewReader("data"))
				m.SetInput(input)

				_, err := m.Copy()
				Expect(err).To(Equal(io.ErrShortWrite))
			})
		})
	})

	Describe("Integration scenarios", func() {
		Context("mixed read, write, and copy operations", func() {
			It("[TC-CP-005] should handle sequential operations", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				// Direct write
				m.Write([]byte("direct: "))

				// Setup input and copy
				input := io.NopCloser(strings.NewReader("copied"))
				m.SetInput(input)
				m.Copy()

				Expect(buf.String()).To(Equal("direct: copied"))
			})

			It("[TC-CP-005] should handle writer changes between operations", func() {
				var buf1 bytes.Buffer
				m.AddWriter(&buf1)

				input := io.NopCloser(strings.NewReader("data1"))
				m.SetInput(input)
				m.Copy()

				// Add another writer and write directly
				var buf2 bytes.Buffer
				m.AddWriter(&buf2)
				m.Write([]byte(" data2"))

				Expect(buf1.String()).To(Equal("data1 data2"))
				Expect(buf2.String()).To(Equal(" data2"))
			})

			It("[TC-CP-005] should handle clean and re-add between copies", func() {
				var buf1 bytes.Buffer
				m.AddWriter(&buf1)

				input1 := io.NopCloser(strings.NewReader("first"))
				m.SetInput(input1)
				m.Copy()

				m.Clean()

				var buf2 bytes.Buffer
				m.AddWriter(&buf2)

				input2 := io.NopCloser(strings.NewReader("second"))
				m.SetInput(input2)
				m.Copy()

				Expect(buf1.String()).To(Equal("first"))
				Expect(buf2.String()).To(Equal("second"))
			})
		})

		Context("using Reader() and Writer() directly", func() {
			It("should allow direct access to reader", func() {
				input := io.NopCloser(strings.NewReader("test"))
				m.SetInput(input)

				reader := m.Reader()
				buf := make([]byte, 4)
				n, err := reader.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(4))
				Expect(string(buf)).To(Equal("test"))
			})

			It("should allow direct access to writer", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				writer := m.Writer()
				n, err := writer.Write([]byte("direct"))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(6))
				Expect(buf.String()).To(Equal("direct"))
			})

			It("should allow manual copy using Reader and Writer", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				input := io.NopCloser(strings.NewReader("manual copy"))
				m.SetInput(input)

				n, err := io.Copy(m.Writer(), m.Reader())
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(int64(11)))
				Expect(buf.String()).To(Equal("manual copy"))
			})
		})
	})
})
