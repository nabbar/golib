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
	"strings"

	libarc "github.com/nabbar/golib/archive"
	arcarc "github.com/nabbar/golib/archive/archive"
	arccmp "github.com/nabbar/golib/archive/compress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testingCompress(alg arccmp.Algorithm, str, ext string) {
	var (
		tmp arccmp.Algorithm
		jsn = []byte("\"" + str + "\"")
		res []byte
	)

	Expect(libarc.ParseCompression(str)).To(Equal(alg))
	Expect(strings.ToLower(alg.String())).To(Equal(str))
	Expect(strings.ToLower(alg.Extension())).To(Equal(ext))
	Expect(alg.IsNone()).To(BeFalse())

	err = tmp.UnmarshalJSON(jsn)
	Expect(err).ToNot(HaveOccurred())
	Expect(tmp).To(Equal(alg))

	res, err = tmp.MarshalJSON()
	Expect(err).ToNot(HaveOccurred())
	Expect(res).To(Equal(jsn))
}

func testingArchive(alg arcarc.Algorithm, str, ext string) {
	var (
		tmp arcarc.Algorithm
		jsn = []byte("\"" + str + "\"")
		res []byte
	)

	Expect(libarc.ParseArchive(str)).To(Equal(alg))
	Expect(strings.ToLower(alg.String())).To(Equal(str))
	Expect(strings.ToLower(alg.Extension())).To(Equal(ext))
	Expect(alg.IsNone()).To(BeFalse())

	err = tmp.UnmarshalJSON(jsn)
	Expect(err).ToNot(HaveOccurred())
	Expect(tmp).To(Equal(alg))

	res, err = tmp.MarshalJSON()
	Expect(err).ToNot(HaveOccurred())
	Expect(res).To(Equal(jsn))
}

var _ = Describe("TC-AL-001: archive/archive/algorithm", func() {
	Context("TC-AL-010: Using algorithm const", func() {
		It("TC-AL-011: gzip must succeed", func() {
			testingCompress(arccmp.Gzip, "gzip", ".gz")
		})
		It("TC-AL-012: bzip2 must succeed", func() {
			testingCompress(arccmp.Bzip2, "bzip2", ".bz2")
		})
		It("TC-AL-013: lz4 must succeed", func() {
			testingCompress(arccmp.LZ4, "lz4", ".lz4")
		})
		It("TC-AL-014: xz must succeed", func() {
			testingCompress(arccmp.XZ, "xz", ".xz")
		})
		It("TC-AL-015: tar must succeed", func() {
			testingArchive(arcarc.Tar, "tar", ".tar")
		})
		It("TC-AL-016: zip must succeed", func() {
			testingArchive(arcarc.Zip, "zip", ".zip")
		})
	})
})
