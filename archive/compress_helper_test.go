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
