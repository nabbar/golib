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

package bufferReadCloser

import (
	"bufio"
	"bytes"
	"io"
)

type FuncClose func() error

type Buffer interface {
	io.Reader
	io.ReaderFrom
	io.ByteReader
	io.RuneReader
	io.Writer
	io.WriterTo
	io.ByteWriter
	io.StringWriter
	io.Closer
}

// @deprecated use NewBuffer instead of New
func New(b *bytes.Buffer) Buffer {
	return NewBuffer(b, nil)
}

func NewBuffer(b *bytes.Buffer, fct FuncClose) Buffer {
	return &buf{
		b: b,
		f: fct,
	}
}

type Reader interface {
	io.Reader
	io.WriterTo
	io.Closer
}

func NewReader(b *bufio.Reader, fct FuncClose) Reader {
	return &rdr{
		b: b,
		f: fct,
	}
}

type Writer interface {
	io.Writer
	io.StringWriter
	io.ReaderFrom
	io.Closer
}

func NewWriter(b *bufio.Writer, fct FuncClose) Writer {
	return &wrt{
		b: b,
		f: fct,
	}
}

type ReadWriter interface {
	Reader
	Writer
}

func NewReadWriter(b *bufio.ReadWriter, fct FuncClose) ReadWriter {
	return &rwt{
		b: b,
		f: fct,
	}
}
