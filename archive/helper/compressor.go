/*
 *  MIT License
 *
 *  Copyright (c) 2024 Salim Amine Bou Aram
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
	"io"
	"sync/atomic"
)

type compressWriter struct {
	dst io.WriteCloser
}

func (o *compressWriter) Read(p []byte) (n int, err error) {
	return 0, ErrInvalidSource
}

func (o *compressWriter) Write(p []byte) (n int, err error) {
	return o.dst.Write(p)
}

func (o *compressWriter) Close() error {
	return o.dst.Close()
}

// compressor handles data compression in chunks.
type compressReader struct {
	src io.ReadCloser
	wrt io.WriteCloser
	buf *bytes.Buffer
	clo *atomic.Bool
}

// Read for compressor compresses the data and reads it from the buffer in chunks.
func (o *compressReader) Read(p []byte) (n int, err error) {
	if o.src == nil {
		return 0, ErrInvalidSource
	}

	var size int

	if s := cap(p); s < chunkSize {
		size = chunkSize
	} else {
		size = s
	}

	if o.clo.Load() && o.buf.Len() == 0 {
		return 0, io.EOF
	}

	if o.buf.Len() < size && !o.clo.Load() {
		if _, err = o.fill(size); err != nil {
			return 0, err
		}
	}

	n, err = o.buf.Read(p)

	if n > 0 {
		return n, nil
	} else if err == nil {
		err = io.EOF
	}

	return 0, err
}

// fill handles compressing data from the source and writing to the buffer.
func (o *compressReader) fill(size int) (n int, err error) {
	var (
		buf    = make([]byte, size)
		errWrt error
		errclo error
	)

	for o.buf.Len() < size {
		if n, err = o.src.Read(buf); err != nil && err != io.EOF {
			return 0, err
		}

		if n > 0 {
			if _, errWrt = o.wrt.Write(buf[:n]); errWrt != nil {
				return 0, errWrt
			}
		}

		if err == io.EOF {
			o.clo.Store(true)

			errWrt = o.wrt.Close()
			errclo = o.src.Close()

			if errclo != nil {
				return 0, errclo
			} else if errWrt != nil {
				return 0, errWrt
			}

			return o.buf.Len(), nil
		} else if err != nil {
			return n, err
		}
	}

	data := o.buf.Bytes()
	o.buf.Reset()

	if _, err = o.buf.Write(data); err != nil {
		return 0, err
	}

	return o.buf.Len(), nil
}

// Close closes the compressor and underlying writer.
func (o *compressReader) Close() (err error) {

	o.clo.Store(true)

	if o.buf != nil {
		o.buf.Reset()
	}

	if o.wrt != nil {
		return o.wrt.Close()
	}

	return nil
}

func (o *compressReader) Write(p []byte) (n int, err error) {
	return 0, ErrInvalidSource
}
