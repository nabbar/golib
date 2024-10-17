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
	"bytes"
	"io"
	"reflect"

	"github.com/nabbar/golib/archive/compress"
	"github.com/nabbar/golib/archive/compress/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compress Helper Test", func() {

	var (
		dcHelper        helper.Helper
		compressionAlgo = []compress.Algorithm{compress.Gzip, compress.Bzip2, compress.LZ4, compress.XZ}
	)

	for _, algo := range compressionAlgo {
		Context(algo.String(), func() {
			var (
				buf                      = make([]byte, 1024)
				compressed, decompressed bytes.Buffer
			)

			BeforeEach(func() {
				// Create the compressor helper
				dcHelper, err = helper.NewHelper(algo)
				Expect(err).NotTo(HaveOccurred())

			})

			It("should compress and then decompress correctly", func() {

				// Initialize the compressor
				err = dcHelper.Compress(bytes.NewReader([]byte(loremIpsum)))
				Expect(err).NotTo(HaveOccurred())

				// Read compressed data
				for {
					n, err := dcHelper.Read(buf)
					if err == io.EOF {
						break
					}
					Expect(err).NotTo(HaveOccurred())
					compressed.Write(buf[:n])
				}

				// Initialize decompressor
				err := dcHelper.Decompress(bytes.NewReader(compressed.Bytes()))
				Expect(err).NotTo(HaveOccurred())

				// Read decompressed data
				for {
					n, err := dcHelper.Read(buf)
					if err == io.EOF {
						break
					}
					Expect(err).NotTo(HaveOccurred())
					decompressed.Write(buf[:n])
				}

				// Check if decompressed data matches the original data
				Expect(reflect.DeepEqual([]byte(loremIpsum), decompressed.Bytes())).To(BeTrue(), "unexpected decompressed data")
			})

			AfterEach(func() {
				compressed.Reset()
				decompressed.Reset()
			})
		})
	}
})
