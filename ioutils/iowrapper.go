/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package ioutils

import "io"

type IOWrapper struct {
	iow   interface{}
	read  func(p []byte) []byte
	write func(p []byte) []byte
}

func NewIOWrapper(ioInput interface{}) *IOWrapper {
	return &IOWrapper{
		iow: ioInput,
	}
}

func (w *IOWrapper) SetWrapper(read func(p []byte) []byte, write func(p []byte) []byte) {
	if read != nil {
		w.read = read
	}
	if write != nil {
		w.write = write
	}
}

func (w IOWrapper) Read(p []byte) (n int, err error) {
	if w.read != nil {
		return w.iow.(io.Reader).Read(w.read(p))
	}

	return w.iow.(io.Reader).Read(p)
}

func (w IOWrapper) Write(p []byte) (n int, err error) {
	if w.write != nil {
		return w.iow.(io.Writer).Write(w.write(p))
	}

	return w.iow.(io.Writer).Write(p)
}

func (w IOWrapper) Seek(offset int64, whence int) (int64, error) {
	return w.iow.(io.Seeker).Seek(offset, whence)
}

func (w IOWrapper) Close() error {
	return w.iow.(io.Closer).Close()
}
