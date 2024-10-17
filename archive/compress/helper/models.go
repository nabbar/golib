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

package helper

import (
	"bytes"
	"fmt"
	"io"

	"github.com/nabbar/golib/archive/compress"
)

// engine manages both compression and decompression operations.
type engine struct {
	operation    string
	compressor   *compressor
	decompressor *decompressor
	algo         compress.Algorithm
}

// compressor handles data compression in chunks.
type compressor struct {
	source io.Reader
	writer io.WriteCloser
	buffer *bytes.Buffer
	closed bool
}

// decompressor handles data decompression in chunks.
type decompressor struct {
	source io.Reader
	writer io.WriteCloser
	buffer *bytes.Buffer
	closed bool
}

// bufferRWCloser wraps a bytes.Buffer to implement io.ReadWriteCloser.
type bufferRWCloser struct {
	*bytes.Buffer
}

func (bc *bufferRWCloser) Close() error {
	return nil
}

// newBufferCloser converts a bytes.Buffer to an io.ReadWriteCloser.
func newBufferRWCloser(buffer *bytes.Buffer) io.ReadWriteCloser {
	return &bufferRWCloser{buffer}
}

// Compress initializes the compressor.
func (e *engine) Compress(source io.Reader) error {
	var buffer bytes.Buffer
	writer, err := e.algo.Writer(newBufferRWCloser(&buffer))
	if err != nil {
		return err
	}

	e.compressor = &compressor{
		source: source,
		writer: writer,
		buffer: &buffer,
		closed: false,
	}

	e.operation = "compress"
	return nil
}

// Decompress initializes the decompressor.
func (e *engine) Decompress(source io.Reader) error {
	var (
		err    error
		buffer bytes.Buffer
		reader io.ReadCloser
	)

	reader, err = e.algo.Reader(source)
	if err != nil {
		return err
	}

	e.decompressor = &decompressor{
		source: reader,
		writer: newBufferRWCloser(&buffer),
		buffer: &buffer,
		closed: false,
	}

	e.operation = "decompress"
	return nil
}

// Read handles reading from the compressor or decompressor.
func (e *engine) Read(p []byte) (int, error) {

	if e.operation == "" {
		return 0, fmt.Errorf("operation mode not set, please call Compress or Decompress first")
	}

	switch e.operation {
	case "compress":
		return e.compressor.Read(p)
	case "decompress":
		return e.decompressor.Read(p)
	default:
		return 0, io.EOF
	}
}

// Close handles closing the compressor or decompressor.
func (e *engine) Close() error {
	switch e.operation {
	case "compress":
		if e.compressor != nil {
			return e.compressor.Close()
		}
	case "decompress":
		if e.decompressor != nil {
			return e.decompressor.Close()
		}
	}
	return nil
}