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
	"bytes"
	"io"
	"reflect"

	arccmp "github.com/nabbar/golib/archive/compress"
	archlp "github.com/nabbar/golib/archive/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compress Helper Test", func() {

	var (
		dcHelper        archlp.Helper
		compressionAlgo = []arccmp.Algorithm{arccmp.Gzip, arccmp.XZ, arccmp.LZ4, arccmp.Bzip2}
	)

	for _, algo := range compressionAlgo {
		Context(algo.String(), func() {

			It("should compress and then decompress correctly in reader mode ", func() {
				var (
					cmpNbr       int64
					decNbr       int64
					compressed   = bytes.NewBuffer(make([]byte, 0))
					decompressed = bytes.NewBuffer(make([]byte, 0))
				)

				// Create the compressor helper
				dcHelper, err = archlp.New(algo, archlp.ReaderMode)
				Expect(err).NotTo(HaveOccurred())
				Expect(dcHelper).NotTo(BeNil())

				// Initialize the compressor
				err = dcHelper.Compress(bytes.NewReader([]byte(loremIpsum)))
				Expect(err).NotTo(HaveOccurred())

				// Read compressed data
				cmpNbr, err = io.Copy(compressed, dcHelper)

				Expect(err).NotTo(HaveOccurred())
				Expect(cmpNbr).To(BeNumerically(">", 0))

				// Initialize decompressor
				err := dcHelper.Decompress(bytes.NewReader(compressed.Bytes()))
				Expect(err).NotTo(HaveOccurred())

				// Read compressed data
				decNbr, err = io.Copy(decompressed, dcHelper)
				Expect(err).NotTo(HaveOccurred())
				Expect(decNbr).To(BeNumerically(">", 0))

				// Check if decompressed data matches the original data
				Expect(reflect.DeepEqual([]byte(loremIpsum), decompressed.Bytes())).To(BeTrue(), "unexpected decompressed data")

			})

			It("should compress and then decompress correctly in writer mode ", func() {
				var (
					n            int
					compressed   = bytes.NewBuffer(make([]byte, 0))
					decompressed = bytes.NewBuffer(make([]byte, 0))
				)

				// Create the compressor helper
				dcHelper, err = archlp.New(algo, archlp.WriterMode)
				Expect(err).NotTo(HaveOccurred())
				Expect(dcHelper).NotTo(BeNil())

				// Initialize the compressor
				err = dcHelper.Compress(compressed)
				Expect(err).NotTo(HaveOccurred())

				// Compress
				n, err = dcHelper.Write([]byte(loremIpsum))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(BeNumerically(">", 0))

				// Initialize decompressor
				err := dcHelper.Decompress(decompressed)
				Expect(err).NotTo(HaveOccurred())

				// decompress
				n, err = dcHelper.Write(compressed.Bytes())
				Expect(err).NotTo(HaveOccurred())

				// Check if decompressed data matches the original data
				Expect(reflect.DeepEqual([]byte(loremIpsum), decompressed.Bytes())).To(BeTrue(), "unexpected decompressed data")

			})

		})
	}
})
