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

// Seek sets the offset for the next Read or Write on the file.
// It implements the io.Seeker interface with integrated reset callback triggering.
// The reset callback is invoked after a successful seek operation.
// whence values: io.SeekStart, io.SeekCurrent, io.SeekEnd.
// Returns the new offset and any error encountered.
func (o *progress) Seek(offset int64, whence int) (int64, error) {
	n, err := o.seek(offset, whence)

	if err != nil {
		o.reset()
	}

	return n, err
}

// seek is an internal helper that performs the actual seek operation
// without triggering callbacks. It wraps os.File.Seek() with nil checks.
// Returns ErrorNilPointer if called on nil instance or closed file.
func (o *progress) seek(offset int64, whence int) (int64, error) {
	if o == nil || o.f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	n, err := o.f.Seek(offset, whence)
	return n, err
}
