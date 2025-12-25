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

package archive

import (
	"errors"
	"io"

	arctar "github.com/nabbar/golib/archive/archive/tar"
	arctps "github.com/nabbar/golib/archive/archive/types"
	arczip "github.com/nabbar/golib/archive/archive/zip"
)

var (
	// ErrInvalidAlgorithm is returned when attempting to create a reader or writer
	// for an unsupported or None algorithm.
	ErrInvalidAlgorithm = errors.New("invalid algorithm")
)

// Reader creates an archive reader for the specified algorithm from the provided io.ReadCloser.
// The reader allows extraction of files from the archive using the types.Reader interface.
//
// Supported algorithms:
//   - Tar: creates a tar.Reader (requires sequential io.Reader)
//   - Zip: creates a zip.Reader (requires io.ReaderAt for random access)
//
// Parameters:
//   - r: the input stream to read the archive from. The caller is responsible for closing
//     this stream when done (the returned reader does not take ownership).
//
// Returns:
//   - arctps.Reader: the archive reader with methods for listing, extracting, and walking files
//   - error: ErrInvalidAlgorithm if algorithm is None or invalid, or an error from the
//     underlying reader creation (e.g., invalid archive format, missing capabilities)
//
// Notes:
//   - Tar readers work with any io.ReadCloser but can only be read sequentially
//   - Zip readers require io.ReaderAt and io.Seeker capabilities for random access
//   - The reader must be closed when done to release resources
func (a Algorithm) Reader(r io.ReadCloser) (arctps.Reader, error) {
	switch a {
	case Tar:
		return arctar.NewReader(r)
	case Zip:
		return arczip.NewReader(r)
	default:
		return nil, ErrInvalidAlgorithm
	}
}

// Writer creates an archive writer for the specified algorithm from the provided io.WriteCloser.
// The writer allows adding files to the archive using the types.Writer interface.
//
// Supported algorithms:
//   - Tar: creates a tar.Writer
//   - Zip: creates a zip.Writer
//
// Parameters:
//   - w: the output stream to write the archive to. The caller is responsible for closing
//     this stream after closing the writer (writer.Close() must be called first to flush data).
//
// Returns:
//   - arctps.Writer: the archive writer with methods for adding files and directories
//   - error: ErrInvalidAlgorithm if algorithm is None or invalid
//
// Usage pattern:
//
//	writer, err := alg.Writer(file)
//	if err != nil {
//	    return err
//	}
//	defer writer.Close()  // Close writer first (flushes internal buffers)
//	// ... add files ...
//	// file will be closed by defer or explicitly after writer.Close()
func (a Algorithm) Writer(w io.WriteCloser) (arctps.Writer, error) {
	switch a {
	case Tar:
		return arctar.NewWriter(w)
	case Zip:
		return arczip.NewWriter(w)
	default:
		return nil, ErrInvalidAlgorithm
	}
}
