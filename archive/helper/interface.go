/*
 *  MIT License
 *
 *  Copyright (c) 2024 Salim Amine BOU ARAM & Nicolas JUHEL
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

const chunkSize = 512

var (
	ErrInvalidSource    = errors.New("invalid source")
	ErrClosedResource   = errors.New("closed resource")
	ErrInvalidOperation = errors.New("invalid operation")
)

type Helper interface {
	io.ReadWriteCloser
}

// New creates a new Helper instance based on the provided algorithm, operation, and source.
// Algo is the compression algorithm to use, which can be one of the predefined algorithms in the arccmp package.
// Operation defines the type of operation to perform with the archive helper. It can be either Compress or Decompress.
// The source can be an io.Reader for reading or an io.Writer for writing.
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
// It uses the specified compression algorithm and operation type (Compress or Decompress).
// The returned Helper can be used to read compressed or decompressed data based on the operation specified.
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
// It uses the specified compression algorithm and operation type (Compress or Decompress).
// The returned Helper can be used to write compressed or decompressed data based on the operation specified.
func NewWriter(algo arccmp.Algorithm, ope Operation, dst io.Writer) (Helper, error) {
	switch ope {
	case Compress:
		return makeCompressWriter(algo, dst)
	case Decompress:
		return makeDeCompressWriter(algo, dst)
	}

	return nil, ErrInvalidOperation
}
