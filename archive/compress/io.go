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

package compress

import (
	"compress/bzip2"
	"compress/gzip"
	"io"

	bz2 "github.com/dsnet/compress/bzip2"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

// Reader wraps the provided io.Reader with a decompression reader for this algorithm.
// The returned io.ReadCloser transparently decompresses data read from the underlying reader.
//
// For None algorithm, it returns io.NopCloser(r) which provides pass-through reading.
// For algorithms that don't provide a Close() method (Bzip2, LZ4, XZ), the reader
// is wrapped with io.NopCloser to satisfy the io.ReadCloser interface.
//
// Supported algorithms:
//   - Gzip:  Returns *gzip.Reader (has native Close method)
//   - Bzip2: Returns io.NopCloser(bzip2.NewReader(r)) (read-only stdlib)
//   - LZ4:   Returns io.NopCloser(lz4.NewReader(r))
//   - XZ:    Returns io.NopCloser(xz.NewReader(r))
//   - None:  Returns io.NopCloser(r) (pass-through)
//
// Returns an error if the decompression reader cannot be created (e.g., invalid header).
//
// Example:
//
//	file, _ := os.Open("data.gz")
//	defer file.Close()
//	reader, err := compress.Gzip.Reader(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
//	data, _ := io.ReadAll(reader)  // Decompressed data
func (a Algorithm) Reader(r io.Reader) (io.ReadCloser, error) {
	switch a {
	case Bzip2:
		return io.NopCloser(bzip2.NewReader(r)), nil
	case Gzip:
		return gzip.NewReader(r)
	case LZ4:
		return io.NopCloser(lz4.NewReader(r)), nil
	case XZ:
		c, e := xz.NewReader(r)
		return io.NopCloser(c), e
	default:
		return io.NopCloser(r), nil
	}
}

// Writer wraps the provided io.WriteCloser with a compression writer for this algorithm.
// The returned io.WriteCloser transparently compresses data written to it before
// passing it to the underlying writer.
//
// IMPORTANT: The returned writer MUST be closed to flush any buffered compressed data.
// Failing to close the writer may result in truncated or corrupted output.
//
// For None algorithm, it returns the original writer unchanged (pass-through).
//
// Supported algorithms:
//   - Gzip:  Returns *gzip.Writer (default compression level)
//   - Bzip2: Returns *bzip2.Writer from dsnet/compress (supports writing)
//   - LZ4:   Returns *lz4.Writer (default options)
//   - XZ:    Returns *xz.Writer (default compression level)
//   - None:  Returns w unchanged (pass-through)
//
// Returns an error if the compression writer cannot be created (e.g., invalid configuration).
//
// Example:
//
//	file, _ := os.Create("output.gz")
//	defer file.Close()
//	writer, err := compress.Gzip.Writer(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer writer.Close()  // MUST close to flush buffers
//	writer.Write([]byte("This will be compressed"))
func (a Algorithm) Writer(w io.WriteCloser) (io.WriteCloser, error) {
	switch a {
	case Bzip2:
		return bz2.NewWriter(w, nil)
	case Gzip:
		return gzip.NewWriter(w), nil
	case LZ4:
		return lz4.NewWriter(w), nil
	case XZ:
		return xz.NewWriter(w)
	default:
		return w, nil
	}
}
