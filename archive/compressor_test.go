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
	"testing"

	"github.com/nabbar/golib/archive/compress"
)

func TestCompressor(t *testing.T) {
	var (
		n               int
		compressor      *compress.Compressor
		minCapacity     = 256
		maxCapacity     = 1024
		compressionAlgo = []compress.Algorithm{compress.Gzip, compress.Bzip2, compress.LZ4, compress.XZ}
	)

	for _, algo := range compressionAlgo {
		var (
			buf            = make([]byte, maxCapacity)
			compressedData bytes.Buffer
			decReader      io.Reader
		)

		dataReader := bytes.NewReader([]byte(loremIpsum))

		// Compress data
		compressor, err = compress.NewCompressor(dataReader, algo, minCapacity, maxCapacity)
		if err != nil {
			t.Fatalf("Failed to create compressor: %v", err)
		}

		// Reading compressed Data
		for {
			n, err = compressor.Read(buf)
			if err == io.EOF && n == 0 {
				break
			}
			if err != nil {
				t.Fatalf("Failed to read compressed data: %v", err)
			}
			compressedData.Write(buf[:n])
		}

		if decReader, err = compressor.Decompress(compressedData.Bytes()); err != nil {
			t.Fatalf("Decompression error: %v", err)
		}

		if !reflect.DeepEqual(streamToByte(decReader), []byte(loremIpsum)) {
			t.Fatalf("the decompressed data is not equal the original expected data")
		}

	}

}

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(stream); err != nil {
		return nil
	}
	return buf.Bytes()
}
