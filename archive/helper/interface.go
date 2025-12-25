/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package helper

import (
	"errors"
	"io"

	arccmp "github.com/nabbar/golib/archive/compress"
)

// chunkSize defines the default buffer size used for internal operations.
const chunkSize = 512

var (
	// ErrInvalidSource is returned when the provided source is not an io.Reader or io.Writer.
	ErrInvalidSource = errors.New("invalid source")
	// ErrClosedResource is returned when attempting to write to a closed resource.
	ErrClosedResource = errors.New("closed resource")
	// ErrInvalidOperation is returned when an unsupported operation is requested.
	ErrInvalidOperation = errors.New("invalid operation")
)

// Helper provides a unified interface for compression and decompression operations.
// It implements io.ReadWriteCloser to enable transparent compression/decompression
// in streaming scenarios.
type Helper interface {
	io.ReadWriteCloser
}

// New creates a new Helper instance based on the provided algorithm, operation, and source.
//
// Parameters:
//   - algo: The compression algorithm to use (from github.com/nabbar/golib/archive/compress)
//   - ope: The operation type (Compress or Decompress)
//   - src: The data source, must be either io.Reader or io.Writer
//
// Returns:
//   - Helper: A new Helper instance for compression/decompression operations
//   - error: ErrInvalidSource if src is neither io.Reader nor io.Writer
//
// The function automatically determines whether to create a reader or writer based
// on the type of src. For io.Reader, it creates a Helper that can be read from.
// For io.Writer, it creates a Helper that can be written to.
func New(algo arccmp.Algorithm, ope Operation, src any) (h Helper, err error) {
	if r, k := src.(io.Reader); k {
		return NewReader(algo, ope, r)
	}
	if w, k := src.(io.Writer); k {
		return NewWriter(algo, ope, w)
	}
	return nil, ErrInvalidSource
}

// NewReader creates a new Helper instance for reading data from the provided io.Reader.
//
// Parameters:
//   - algo: The compression algorithm to use
//   - ope: The operation type (Compress to compress while reading, Decompress to decompress while reading)
//   - src: The source reader to read data from
//
// Returns:
//   - Helper: A new Helper instance that wraps the source reader
//   - error: ErrInvalidOperation if the operation is not Compress or Decompress
//
// When operation is Compress, reading from the returned Helper will compress data from src.
// When operation is Decompress, reading from the returned Helper will decompress data from src.
func NewReader(algo arccmp.Algorithm, ope Operation, src io.Reader) (Helper, error) {
	switch ope {
	case Compress:
		return makeCompressReader(algo, src)
	case Decompress:
		return makeDeCompressReader(algo, src)
	}

	return nil, ErrInvalidOperation
}

// NewWriter creates a new Helper instance for writing data to the provided io.Writer.
//
// Parameters:
//   - algo: The compression algorithm to use
//   - ope: The operation type (Compress to compress while writing, Decompress to decompress while writing)
//   - dst: The destination writer to write data to
//
// Returns:
//   - Helper: A new Helper instance that wraps the destination writer
//   - error: ErrInvalidOperation if the operation is not Compress or Decompress
//
// When operation is Compress, writing to the returned Helper will compress data to dst.
// When operation is Decompress, writing to the returned Helper will decompress data to dst.
func NewWriter(algo arccmp.Algorithm, ope Operation, dst io.Writer) (Helper, error) {
	switch ope {
	case Compress:
		return makeCompressWriter(algo, dst)
	case Decompress:
		return makeDeCompressWriter(algo, dst)
	}

	return nil, ErrInvalidOperation
}
