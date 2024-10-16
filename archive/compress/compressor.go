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

package compress

import (
	"bytes"
	"fmt"
	"io"
)

type Compressor struct {
	source      io.Reader
	writer      io.WriteCloser
	algo        Algorithm
	buffer      *bytes.Buffer
	minCapacity int
	maxCapacity int
	closed      bool
}

type bufferWriteCloser struct {
	*bytes.Buffer
}

func (bwc *bufferWriteCloser) Close() error {
	return nil
}

// NewCompressor creates a new compressor
func NewCompressor(source io.Reader, algo Algorithm, minCapacity,
	maxCapacity int) (*Compressor, error) {
	var buffer bytes.Buffer

	writer, err := algo.Writer(&bufferWriteCloser{&buffer})

	if err != nil {
		return nil, err
	}

	return &Compressor{
		source:      source,
		writer:      writer,
		algo:        algo,
		buffer:      &buffer,
		minCapacity: minCapacity,
		maxCapacity: maxCapacity,
		closed:      false,
	}, nil
}

func (c *Compressor) Read(outputBuffer []byte) (int, error) {

	if c.closed && c.buffer.Len() == 0 {
		return 0, io.EOF
	}

	if c.buffer.Len() == 0 {

		capacity := cap(outputBuffer)

		if capacity < c.minCapacity {
			capacity = c.minCapacity
		}

		if capacity > c.maxCapacity {
			capacity = c.maxCapacity
		}

		tempBuffer := make([]byte, capacity)

		n, err := c.source.Read(tempBuffer)
		if err != nil && err != io.EOF {
			return 0, err
		}

		if n > 0 {
			if _, err = c.writer.Write(tempBuffer[:n]); err != nil {
				return 0, err
			}
		}

		if err == io.EOF {
			if closeErr := c.writer.Close(); closeErr != nil {
				return 0, closeErr
			}
			c.closed = true
		}

		compressedData := c.buffer.Bytes()
		c.buffer.Reset()

		if _, err = c.buffer.Write(compressedData); err != nil {
			return 0, err
		}
	}

	n, err := c.buffer.Read(outputBuffer)
	if err == io.EOF && c.buffer.Len() == 0 {
		err = nil
	}

	return n, err
}

func (c *Compressor) Decompress(data []byte) (io.Reader, error) {

	if c == nil {
		return bytes.NewBuffer(data), nil
	}

	reader, err := c.algo.Reader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create a reader for data decompressing: %w", err)
	}

	defer func(reader io.ReadCloser) {
		err = reader.Close()
		if err != nil {

		}
	}(reader)

	return reader, nil
}
