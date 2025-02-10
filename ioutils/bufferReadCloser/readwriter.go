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
	"io"
)

type rwt struct {
	b *bufio.ReadWriter
	f FuncClose
}

func (b *rwt) Read(p []byte) (n int, err error) {
	return b.b.Read(p)
}

func (b *rwt) WriteTo(w io.Writer) (n int64, err error) {
	return b.b.WriteTo(w)
}

func (b *rwt) ReadFrom(r io.Reader) (n int64, err error) {
	return b.b.ReadFrom(r)
}

func (b *rwt) Write(p []byte) (n int, err error) {
	return b.b.Write(p)
}

func (b *rwt) WriteString(s string) (n int, err error) {
	return b.b.WriteString(s)
}

func (b *rwt) Close() error {
	_ = b.b.Flush()

	// cannot call Reset => ambigous func can be Reader and Writer
	// b.b.Reset(nil)

	if b.f != nil {
		return b.f()
	}

	return nil
}
