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
	"io"
	"io/fs"
	"os"

	libarc "github.com/nabbar/golib/archive"
	arcarc "github.com/nabbar/golib/archive/archive"
	arctps "github.com/nabbar/golib/archive/archive/types"
	arccmp "github.com/nabbar/golib/archive/compress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("archive/tar+gzip with extract all", func() {
	Context("Create a tar+gzip archive file and extract it with extract all", func() {
		It("Create a tar archive must succeed", func() {
			var (
				hdf *os.File
				gzp io.WriteCloser
				wrt arctps.Writer
			)

			defer func() {
				if hdf != nil {
					_ = hdf.Close()
				}
			}()

			arc[arcarc.Tar.String()+arccmp.Gzip.String()] = "lorem_ipsum" + arcarc.Tar.Extension() + arccmp.Gzip.Extension()
			hdf, err = os.Create(arc[arcarc.Tar.String()+arccmp.Gzip.String()])
			Expect(err).ToNot(HaveOccurred())
			Expect(hdf).ToNot(BeNil())

			gzp, err = arccmp.Gzip.Writer(hdf)
			Expect(err).ToNot(HaveOccurred())
			Expect(gzp).ToNot(BeNil())

			wrt, err = arcarc.Tar.Writer(gzp)
			Expect(err).ToNot(HaveOccurred())
			Expect(wrt).ToNot(BeNil())

			for f, p := range lst {
				var (
					i fs.FileInfo
					h *os.File
				)

				i, err = os.Stat(f)
				Expect(err).ToNot(HaveOccurred())
				Expect(i).ToNot(BeNil())

				h, err = os.Open(f)
				Expect(err).ToNot(HaveOccurred())
				Expect(h).ToNot(BeNil())

				err = wrt.Add(i, h, p, "")
				Expect(err).ToNot(HaveOccurred())

				err = h.Close()
				Expect(err).To(HaveOccurred())
			}

			err = hdf.Sync()
			Expect(err).ToNot(HaveOccurred())

			err = wrt.Close()
			Expect(err).ToNot(HaveOccurred())

			err = gzp.Close()
			Expect(err).ToNot(HaveOccurred())

			err = hdf.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("Detect all a tar+gzip archive must succeed", func() {
			var hdf *os.File

			defer func() {
				if hdf != nil {
					_ = hdf.Close()
				}
			}()

			hdf, err = os.Open(arc[arcarc.Tar.String()+arccmp.Gzip.String()])
			Expect(err).ToNot(HaveOccurred())
			Expect(hdf).ToNot(BeNil())

			err = libarc.ExtractAll(hdf, arc[arcarc.Tar.String()+arccmp.Gzip.String()], dst)
			Expect(err).ToNot(HaveOccurred())

			err = hdf.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
