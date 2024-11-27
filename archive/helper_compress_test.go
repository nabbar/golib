/*
 *  MIT License
 *
 *  Copyright (c) 2024 Salim Amine Bou Aram
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
	"bytes"
	"io"
	"time"

	arccmp "github.com/nabbar/golib/archive/compress"
	archlp "github.com/nabbar/golib/archive/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compress Helper Test", func() {
	for _, algo := range arccmp.List() {
		Context("For the algo '"+algo.String()+"', in reader mode", func() {
			It("should compress/decompress in embedded stream correctly", func() {
				var (
					siz = len(loremIpsum)
					src = bytes.NewReader([]byte(loremIpsum)) // source
					res = bytes.NewBuffer(make([]byte, 0))    // decompressed
				)

				// init new reader for the algo as compressor
				c, e := archlp.New(algo, archlp.Compress, src)
				Expect(e).NotTo(HaveOccurred())
				Expect(c).NotTo(BeNil())

				// init new reader for the algo as decompressor
				d, e := archlp.New(algo, archlp.Decompress, c)
				Expect(e).NotTo(HaveOccurred())
				Expect(d).NotTo(BeNil())

				// copy res data into buffer
				n, e := io.Copy(res, d)
				Expect(e).NotTo(HaveOccurred())
				Expect(n).To(BeNumerically(">", 0))
				Expect(n).To(BeNumerically("==", siz))

				// closing reader
				Expect(d.Close()).NotTo(HaveOccurred())
				Expect(c.Close()).NotTo(HaveOccurred())

				// check res must be same source
				r := res.Bytes()
				Expect(len(r)).To(BeNumerically("==", siz))
				Expect(r).To(Equal([]byte(loremIpsum)))
			})
		})
		Context("For the algo '"+algo.String()+"', in writer mode", func() {
			It("should compress/decompress in embedded stream correctly", func() {
				var (
					siz = len(loremIpsum)
					src = bytes.NewReader([]byte(loremIpsum)) // source
					res = bytes.NewBuffer(make([]byte, 0))    // decompressed
					wrt = bufio.NewWriter(res)
					tmp *bufio.Writer
				)

				// init new reader for the algo as compressor
				d, e := archlp.New(algo, archlp.Decompress, wrt)
				Expect(e).NotTo(HaveOccurred())
				Expect(d).NotTo(BeNil())
				tmp = bufio.NewWriter(d)

				// init new reader for the algo as compressor
				c, e := archlp.New(algo, archlp.Compress, tmp)
				Expect(e).NotTo(HaveOccurred())
				Expect(c).NotTo(BeNil())

				// copy res data into buffer
				n, e := io.Copy(c, src)
				Expect(e).NotTo(HaveOccurred())
				Expect(n).To(BeNumerically(">", 0))
				Expect(n).To(BeNumerically("==", siz))

				// closing writer compressor and flush to decompressor
				Expect(c.Close()).NotTo(HaveOccurred())
				Expect(tmp.Flush()).To(Succeed())
				Expect(d.Close()).NotTo(HaveOccurred())
				Expect(wrt.Flush()).To(Succeed())

				//need sleep to allow flush by applied on destination
				time.Sleep(time.Second)

				// check res must be same source
				r := res.Bytes()
				s := []byte(loremIpsum)
				Expect(len(r)).To(BeNumerically("==", len(s)))
				Expect(r).To(Equal(s))
			})
		})
	}
})
