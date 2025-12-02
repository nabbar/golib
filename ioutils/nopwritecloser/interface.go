/*
 *  MIT License
 *
 *  Copyright (c) 2025 Nicolas JUHEL
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

package nopwritecloser

import "io"

// New wraps an io.Writer to implement io.WriteCloser with a no-op Close() method.
//
// The returned WriteCloser delegates all Write() calls to the underlying writer
// and implements Close() as a no-operation that always returns nil. This is useful
// when you have an io.Writer but need an io.WriteCloser interface, such as when
// working with APIs that require closeable writers.
//
// Parameters:
//   - w: The io.Writer to wrap
//
// Returns:
//   - io.WriteCloser: A wrapper that adds no-op close semantics
//
// The wrapper is safe for concurrent use if the underlying writer is thread-safe.
// Calling Close() multiple times is safe and will always return nil.
//
// Example:
//
//	var buf bytes.Buffer
//	wc := nopwritecloser.New(&buf)
//	wc.Write([]byte("data"))
//	wc.Close() // No-op, returns nil
func New(w io.Writer) io.WriteCloser {
	return &wrp{w: w}
}
