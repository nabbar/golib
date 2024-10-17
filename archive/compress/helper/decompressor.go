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

// Read for decompressor reads decompressed data from the buffer in chunks.
func (d *decompressor) Read(outputBuffer []byte) (n int, err error) {
	if d.closed && d.buffer.Len() == 0 {
		return 0, io.EOF
	}

	if d.buffer.Len() == 0 {
		if _, err = d.fill(); err != nil {
			return 0, err
		}
	}

	if n, err = d.buffer.Read(outputBuffer); err == io.EOF && d.buffer.Len() == 0 {
		return n, nil
	}

	return n, err

}

// fill handles reading from the decompressed data.
func (d *decompressor) fill() (n int, err error) {
	var tempBuffer = make([]byte, ChunkSize)

	if n, err = d.source.Read(tempBuffer); err != nil && err != io.EOF {
		return 0, err
	}

	if n > 0 {
		if _, err = d.writer.Write(tempBuffer[:n]); err != nil {
			return 0, err
		}
	}

	if err == io.EOF {
		if closeErr := d.writer.Close(); closeErr != nil {
			return 0, closeErr
		}
		d.closed = true
	}

	data := d.buffer.Bytes()
	d.buffer.Reset()

	if _, err = d.buffer.Write(data); err != nil {
		return 0, err
	}

	return n, nil
}

// Close closes the decompressor and underlying writer.
func (d *decompressor) Close() error {
	d.closed = true
	return d.writer.Close()
}
