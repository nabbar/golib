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

import "io"

// Read for compressor compresses the data and reads it from the buffer in chunks.
func (c *compressor) Read(outputBuffer []byte) (n int, err error) {

	if c.closed && c.buffer.Len() == 0 {
		return 0, io.EOF
	}

	if c.buffer.Len() == 0 {
		if _, err = c.fill(); err != nil {
			return 0, err
		}
	}

	if n, err = c.buffer.Read(outputBuffer); err == io.EOF && c.buffer.Len() == 0 {
		return n, nil
	}

	return n, err

}

// fill handles compressing data from the source and writing to the buffer.
func (c *compressor) fill() (n int, err error) {

	var tempBuffer = make([]byte, ChunkSize)

	if n, err = c.source.Read(tempBuffer); err != nil && err != io.EOF {
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

	data := c.buffer.Bytes()

	c.buffer.Reset()

	if _, err = c.buffer.Write(data); err != nil {
		return 0, err
	}

	return n, nil
}

// Close closes the compressor and underlying writer.
func (c *compressor) Close() error {
	c.closed = true
	c.buffer.Reset()
	return c.writer.Close()
}
