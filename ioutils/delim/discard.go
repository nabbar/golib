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

package delim

// DiscardCloser is a no-op implementation of io.ReadWriteCloser that discards all data.
//
// It provides the following behavior:
//   - Read always returns 0 bytes read with no error (acts as immediate EOF)
//   - Write always returns the length of input as successfully written (data is discarded)
//   - Close always returns nil (no-op)
//
// This is useful for:
//   - Testing scenarios where you need a valid io.ReadWriteCloser but don't care about the data
//   - Placeholder implementations where data should be ignored
//   - Benchmarking or profiling where you want to isolate I/O operations
//   - Creating no-op readers/writers for configuration or mocking purposes
//
// Similar to io.Discard (which only implements io.Writer), but DiscardCloser also
// implements io.Reader and io.Closer, making it compatible with BufferDelim.New().
//
// Example:
//
//	// Use as a no-op reader for testing
//	dc := delim.DiscardCloser{}
//	bd := delim.New(dc, '\n', 0, false)
//	defer bd.Close()
//
//	// Any reads will return 0 bytes
//	data, err := bd.ReadBytes()  // Returns (nil, io.EOF)
//
// Example with Writer:
//
//	dc := delim.DiscardCloser{}
//	n, err := dc.Write([]byte("data to discard"))
//	// n == 15, err == nil, but data is not stored anywhere
type DiscardCloser struct{}

// Read implements io.Reader but always returns 0 bytes read.
// This makes DiscardCloser act as an immediate EOF reader.
//
// The provided buffer p is not modified.
//
// Returns:
//   - n: Always 0
//   - err: Always nil
func (d DiscardCloser) Read(p []byte) (n int, err error) {
	return 0, nil
}

// Write implements io.Writer but discards all data.
// It always reports that all bytes were successfully written.
//
// Returns:
//   - n: The length of p (all bytes "written")
//   - err: Always nil
func (d DiscardCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Close implements io.Closer but performs no operation.
//
// Returns:
//   - error: Always nil
func (d DiscardCloser) Close() error {
	return nil
}
