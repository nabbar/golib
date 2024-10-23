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
	"errors"
	"io"
	"sync/atomic"
)

func (e *engine) Close() error {
	return nil
}

type writeCloser struct {
	io.Writer
}

func (w *writeCloser) Close() error {
	return nil
}

func newWCloser(w io.Writer) io.WriteCloser {
	return &writeCloser{Writer: w}
}

type engine struct {
	algo      Algorithm
	buffer    *bytes.Buffer
	writer    io.WriteCloser
	reader    io.Reader
	state     *atomic.Bool
	operation operation
	closed    *atomic.Bool
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

// SetReader configures the engine for compression or decompression via the Read method.
func (e *engine) SetReader(r io.Reader) error {

	if e.state.Load() {
		return errors.New("operation already set")
	}

	switch e.operation {

	case Compress:
		writer, err := e.algo.Writer(newBufferRWCloser(e.buffer))

		if err != nil {
			return err
		}

		e.writer = writer

		e.reader = r

	case Decompress:

		reader, err := e.algo.Reader(r)

		if err != nil {
			return err
		}
		e.reader = reader

	default:
		return errors.New("invalid operation")
	}

	e.state.Store(true)

	return nil
}

// SetWriter configures the engine for compression or decompression via the Write method.
func (e *engine) SetWriter(w io.Writer) error {

	if e.state.Load() {
		return errors.New("operation already set")
	}

	switch e.operation {

	case Compress:

		wc, ok := w.(io.WriteCloser)

		if !ok {
			wc = newWCloser(w)
		}

		cw, err := e.algo.Writer(wc)

		if err != nil {
			return err
		}

		e.writer = cw

	case Decompress:

		e.buffer.Reset()
		e.writer = newWCloser(w)

	default:
		return errors.New("invalid operation")
	}

	e.state.Store(true)

	return nil
}

// SetWriter configures the engine for compression or decompression via the Write method.

func (e *engine) Read(p []byte) (n int, err error) {

	if e.operation == Decompress {

		return e.reader.Read(p)

	} else {

		if e.closed.Load() && e.buffer.Len() == 0 {
			return 0, io.EOF
		}

		if e.buffer.Len() == 0 {

			if _, err = e.fill(); err != nil {
				return 0, err
			}
		}

		if n, err = e.buffer.Read(p); err == io.EOF && e.buffer.Len() == 0 {
			return n, nil
		}

		return n, err
	}

}

func (e *engine) Write(p []byte) (n int, err error) {

	var rc io.ReadCloser

	if e.operation == Compress {

		return e.writer.Write(p)

	} else {

		rc, err = e.algo.Reader(bytes.NewReader(p))

		if err != nil {
			return 0, err
		}

		for {

			buf := make([]byte, chunkSize)

			n, err = rc.Read(buf)

			if err != nil {
				if err == io.EOF {
					break
				}
				return n, err
			}

			if n > 0 {
				if _, err = e.writer.Write(buf[:n]); err != nil {
					return n, err
				}
			}
		}

	}

	return n, nil
}
