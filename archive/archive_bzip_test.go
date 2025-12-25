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
	"bufio"
	"io"
	"os"

	libarc "github.com/nabbar/golib/archive"
	arccmp "github.com/nabbar/golib/archive/compress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TC-BZ-001: archive/compress/bzip", func() {
	Context("TC-BZ-010: Write/Read a bzip compressed file", func() {
		It("TC-BZ-011: Create a bzip compressed file must succeed", func() {
			var (
				hdf *os.File
				buf *bufio.Writer
				wrt io.WriteCloser
			)

			defer func() {
				if hdf != nil {
					_ = hdf.Close()
				}
			}()

			arc[arccmp.Bzip2.String()] = "lorem_ipsum" + arccmp.Bzip2.Extension()

			hdf, err = os.Create(arc[arccmp.Bzip2.String()])
			Expect(err).ToNot(HaveOccurred())
			Expect(hdf).ToNot(BeNil())

			wrt, err = arccmp.Bzip2.Writer(hdf)
			Expect(err).ToNot(HaveOccurred())
			Expect(wrt).ToNot(BeNil())

			buf = bufio.NewWriter(wrt)
			_, err = buf.WriteString(loremIpsum)
			Expect(err).ToNot(HaveOccurred())

			err = buf.Flush()
			Expect(err).ToNot(HaveOccurred())

			err = hdf.Sync()
			Expect(err).ToNot(HaveOccurred())

			err = wrt.Close()
			Expect(err).ToNot(HaveOccurred())

			err = hdf.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-BZ-012: Detect and Extract a bzip compressed file must succeed", func() {
			var (
				hdf *os.File
				alg arccmp.Algorithm
				rdr io.ReadCloser
				buf *bufio.Reader
			)

			defer func() {
				if hdf != nil {
					_ = hdf.Close()
				}
			}()

			hdf, err = os.Open(arc[arccmp.Bzip2.String()])
			Expect(err).ToNot(HaveOccurred())
			Expect(hdf).ToNot(BeNil())

			alg, rdr, err = libarc.DetectCompression(hdf)
			Expect(err).ToNot(HaveOccurred())
			Expect(rdr).ToNot(BeNil())
			Expect(alg).To(Equal(arccmp.Bzip2))

			buf = bufio.NewReader(rdr)
			_, err = io.Copy(io.Discard, buf)
			Expect(err).ToNot(HaveOccurred())

			err = buf.UnreadByte()
			Expect(err).To(HaveOccurred())

			err = rdr.Close()
			Expect(err).ToNot(HaveOccurred())

			err = hdf.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
