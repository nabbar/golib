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
	"path/filepath"

	libarc "github.com/nabbar/golib/archive"
	arcarc "github.com/nabbar/golib/archive/archive"
	arctps "github.com/nabbar/golib/archive/archive/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("archive/archive/tar", func() {
	Context("Write/Read a tar archive file", func() {
		It("Create a tar archive must succeed", func() {
			var (
				hdf *os.File
				wrt arctps.Writer
			)

			defer func() {
				if hdf != nil {
					_ = hdf.Close()
				}
			}()

			arc[arcarc.Tar.String()] = "lorem_ipsum" + arcarc.Tar.Extension()

			hdf, err = os.Create(arc[arcarc.Tar.String()])
			Expect(err).ToNot(HaveOccurred())
			Expect(hdf).ToNot(BeNil())

			wrt, err = arcarc.Tar.Writer(hdf)
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

			err = hdf.Close()
			Expect(err).To(HaveOccurred())
		})

		It("Detect and Extract a tar archive must succeed", func() {
			var (
				hdf *os.File
				alg arcarc.Algorithm
				rdr arctps.Reader
				fnd []string
			)

			defer func() {
				if hdf != nil {
					_ = hdf.Close()
				}
			}()

			hdf, err = os.Open(arc[arcarc.Tar.String()])
			Expect(err).ToNot(HaveOccurred())
			Expect(hdf).ToNot(BeNil())

			alg, rdr, _, err = libarc.DetectArchive(hdf)
			Expect(err).ToNot(HaveOccurred())
			Expect(rdr).ToNot(BeNil())
			Expect(alg).To(Equal(arcarc.Tar))

			fnd, err = rdr.List()
			Expect(err).ToNot(HaveOccurred())
			Expect(fnd).ToNot(BeNil())

			for _, g := range fnd {
				f, ok := lst[filepath.Base(g)]
				Expect(ok).To(BeTrue())
				Expect(g).To(Equal(f))
			}

			for _, f := range lst {
				var (
					i fs.FileInfo
					r io.ReadCloser
					n int64
				)
				Expect(rdr.Has(f)).To(BeTrue())

				i, err = rdr.Info(f)
				Expect(err).ToNot(HaveOccurred())
				Expect(i).ToNot(BeNil())

				r, err = rdr.Get(f)
				Expect(err).ToNot(HaveOccurred())
				Expect(r).ToNot(BeNil())

				n, err = io.Copy(io.Discard, r)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(BeEquivalentTo(i.Size()))

				err = r.Close()
				Expect(err).ToNot(HaveOccurred())
			}

			err = rdr.Close()
			Expect(err).ToNot(HaveOccurred())

			err = hdf.Close()
			Expect(err).To(HaveOccurred())
		})

		It("Detect and Extract a tar archive with walk must succeed", func() {
			var (
				hdf *os.File
				alg arcarc.Algorithm
				rdr arctps.Reader
			)

			defer func() {
				if hdf != nil {
					_ = hdf.Close()
				}
			}()

			hdf, err = os.Open(arc[arcarc.Tar.String()])
			Expect(err).ToNot(HaveOccurred())
			Expect(hdf).ToNot(BeNil())

			alg, rdr, _, err = arcarc.Detect(hdf)
			Expect(err).ToNot(HaveOccurred())
			Expect(rdr).ToNot(BeNil())
			Expect(alg).To(Equal(arcarc.Tar))

			rdr.Walk(func(i fs.FileInfo, r io.ReadCloser, f, t string) bool {
				g, ok := lst[filepath.Base(f)]
				Expect(ok).To(BeTrue())
				Expect(f).To(Equal(g))

				Expect(i).ToNot(BeNil())
				Expect(r).ToNot(BeNil())

				var n int64
				n, err = io.Copy(io.Discard, r)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(BeEquivalentTo(i.Size()))

				err = r.Close()
				Expect(err).ToNot(HaveOccurred())

				return true
			})

			err = rdr.Close()
			Expect(err).ToNot(HaveOccurred())

			err = hdf.Close()
			Expect(err).To(HaveOccurred())
		})
	})
})
