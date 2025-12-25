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
	"bytes"
	"io"
	"sync/atomic"

	arccmp "github.com/nabbar/golib/archive/compress"
	iotnwc "github.com/nabbar/golib/ioutils/nopwritecloser"
)

// makeCompressWriter creates a compression writer that wraps the provided writer.
// Data written to the returned Helper will be compressed using the specified algorithm
// and written to the destination writer.
func makeCompressWriter(algo arccmp.Algorithm, src io.Writer) (h Helper, err error) {
	wc, ok := src.(io.WriteCloser)

	if !ok {
		wc = iotnwc.New(src)
	}

	if wc, err = algo.Writer(wc); err != nil {
		return nil, err
	} else {
		return &compressWriter{
			dst: wc,
		}, nil
	}
}

// compressWriter implements Helper for compression write operations.
// It compresses data as it is written and forwards the compressed data to the underlying writer.
type compressWriter struct {
	dst io.WriteCloser
}

// Read is not supported for compression writers.
// Always returns ErrInvalidSource since compression writers are write-only.
func (o *compressWriter) Read(p []byte) (n int, err error) {
	return 0, ErrInvalidSource
}

// Write compresses the provided data and writes it to the underlying writer.
func (o *compressWriter) Write(p []byte) (n int, err error) {
	return o.dst.Write(p)
}

// Close finalizes the compression stream and closes the underlying writer.
func (o *compressWriter) Close() error {
	return o.dst.Close()
}

// makeCompressReader creates a compression reader that wraps the provided reader.
// Data read from the returned Helper will be compressed on-the-fly from the source reader.
func makeCompressReader(algo arccmp.Algorithm, src io.Reader) (h Helper, err error) {
	rc, ok := src.(io.ReadCloser)

	if !ok {
		rc = io.NopCloser(src)
	}

	var (
		buf = bytes.NewBuffer(make([]byte, 0))
		wrt io.WriteCloser
	)

	wrt, err = algo.Writer(iotnwc.New(buf))

	return &compressReader{
		src: rc,
		wrt: wrt,
		buf: buf,
		clo: new(atomic.Bool),
	}, err
}

// compressReader implements Helper for compression read operations.
// It reads data from the source, compresses it, and provides the compressed data through Read().
type compressReader struct {
	src io.ReadCloser
	wrt io.WriteCloser
	buf *bytes.Buffer
	clo *atomic.Bool
}

// Read compresses data from the source and returns compressed chunks.
// It maintains an internal buffer to handle compression output efficiently.
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

// fill reads data from the source, compresses it, and fills the internal buffer.
// It ensures at least 'size' bytes are available in the buffer before returning.
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

// Close finalizes the compression stream and releases resources.
// It closes the compression writer and resets the internal buffer.
func (o *compressReader) Close() (err error) {
	a := o.clo.Swap(true)

	if o.buf != nil {
		o.buf.Reset()
	}

	if o.wrt != nil && !a {
		return o.wrt.Close()
	}

	return nil
}

// Write is not supported for compression readers.
// Always returns ErrInvalidSource since compression readers are read-only.
func (o *compressReader) Write(p []byte) (n int, err error) {
	return 0, ErrInvalidSource
}
