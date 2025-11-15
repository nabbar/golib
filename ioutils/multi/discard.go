/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

// DiscardCloser is an io.ReadWriteCloser that performs no-op operations.
//
// It implements:
//   - io.Reader: Read always returns (0, nil) without reading any data
//   - io.Writer: Write returns (len(p), nil) without writing any data
//   - io.Closer: Close always returns nil
//
// DiscardCloser is used as a safe default input source in Multi instances
// to prevent nil pointer panics. It's similar to io.Discard but also
// implements io.ReadCloser.
//
// This type is exported for advanced use cases where a no-op ReadWriteCloser
// is needed, though most users should not need to instantiate it directly.
type DiscardCloser struct{}

// Read implements io.Reader by returning zero bytes without error.
// This allows DiscardCloser to satisfy the io.Reader interface while
// performing no actual read operation.
//
// The buffer p is not modified. This method always returns (0, nil).
func (d DiscardCloser) Read(p []byte) (n int, err error) {
	return 0, nil
}

// Write implements io.Writer by accepting and discarding all data.
// It returns len(p) to indicate that all bytes were "written", but
// no data is actually stored or processed.
//
// This method always returns (len(p), nil).
func (d DiscardCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Close implements io.Closer with a no-op operation.
// Since DiscardCloser maintains no state and allocates no resources,
// Close simply returns nil without doing anything.
func (d DiscardCloser) Close() error {
	return nil
}
