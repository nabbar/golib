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

package progress

import (
	"os"
)

// clean resets internal file handles and returns the provided error.
// This is called internally during Close and CloseDelete operations
// to ensure resources are properly released.
func (o *progress) clean(e error) error {
	if o == nil {
		return nil
	}

	o.r = nil
	o.f = nil

	return e
}

// Close closes the file and releases associated resources.
// It implements the io.Closer interface.
// For temporary files (IsTemp() == true), the file is automatically deleted.
// Both the file handle and os.Root are closed if present.
// Returns the first error encountered during closing operations.
func (o *progress) Close() error {
	if o == nil {
		return nil
	}

	var e error

	if o.f != nil {
		e = o.f.Close()
	}

	if o.r != nil {
		if er := o.r.Close(); er != nil && e == nil {
			e = er
		}
	}

	return o.clean(e)
}

// CloseDelete closes the file and then deletes it from the filesystem.
// This is useful for temporary files or when the file is no longer needed.
// The file is removed using os.Root.Remove() if available, otherwise os.Remove().
// Returns the first error encountered during close or delete operations.
func (o *progress) CloseDelete() error {
	if o == nil {
		return nil
	}

	var (
		e error
		n = o.Path()
	)

	if o.f != nil {
		e = o.f.Close()
	}

	if e != nil {
		if o.r != nil {
			_ = o.r.Close()
		}
		return o.clean(e)
	}

	if len(n) < 1 {
		return nil
	}

	if o.r != nil {
		e = o.r.Remove(n)

		if er := o.r.Close(); er != nil && e == nil {
			e = er
		}
	} else {
		e = os.Remove(n)
	}

	return o.clean(e)
}
