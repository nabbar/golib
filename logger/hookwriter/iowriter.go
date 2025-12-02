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

package hookwriter

import (
	"fmt"
)

// Write implements io.Writer, writing formatted log data to the underlying writer.
//
// This method is called internally by Fire() after entry formatting. It delegates
// to the configured io.Writer instance. If the writer is nil (should never happen
// in normal usage), returns an error.
//
// Parameters:
//   - p: Byte slice containing the formatted log entry to write
//
// Returns:
//   - n: Number of bytes written
//   - err: Error from underlying writer, or "writer not setup" if writer is nil
func (o *hkstd) Write(p []byte) (n int, err error) {
	if o.w == nil {
		return 0, fmt.Errorf("logrus.hookstd: writer not setup")
	}

	return o.w.Write(p)
}

// Close implements io.Closer as a no-op.
//
// This hook does not manage the lifecycle of the underlying writer. The caller
// is responsible for closing the writer when appropriate. This method exists
// to satisfy interface requirements and always returns nil.
func (o *hkstd) Close() error {
	return nil
}
