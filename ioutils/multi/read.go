/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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
 *
 */

package multi

import "io"

// readerWrapper wraps an io.Reader to maintain type consistency in atomic.Value.
//
// atomic.Value requires that all stored values have the same concrete type.
// By wrapping all readers in readerWrapper, we ensure this constraint is met
// even when different io.ReadCloser implementations are used.
type readerWrapper struct {
	io.Reader
}

// Read delegates to the wrapped Reader.
func (w *readerWrapper) Read(p []byte) (n int, err error) {
	return w.Reader.Read(p)
}

// Close attempts to close the wrapped Reader if it implements io.Closer.
// Returns nil if the Reader does not implement io.Closer.
func (w *readerWrapper) Close() error {
	if c, k := w.Reader.(io.Closer); k {
		return c.Close()
	}
	return nil
}

// newReadWrapper creates a new readerWrapper. If r is nil, returns a wrapper
// containing DiscardCloser as a safe default.
func newReadWrapper(r io.Reader) *readerWrapper {
	if r == nil {
		return &readerWrapper{
			DiscardCloser{},
		}
	}

	return &readerWrapper{
		Reader: r,
	}
}
