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
	"errors"
	"io"
)

// Write writes len(p) bytes from p to the underlying file.
// It implements the io.Writer interface with integrated progress tracking.
// The increment callback is invoked with the number of bytes written.
// Returns ErrorNilPointer if called on nil instance or closed file.
func (o *progress) Write(p []byte) (n int, err error) {
	if o == nil || o.f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	return o.analyze(o.f.Write(p))
}

// WriteAt writes len(p) bytes from p to the file starting at offset off.
// It implements the io.WriterAt interface with integrated progress tracking.
// The increment callback is invoked with the number of bytes written.
// Returns ErrorNilPointer if called on nil instance or closed file.
func (o *progress) WriteAt(p []byte, off int64) (n int, err error) {
	if o == nil || o.f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	return o.analyze(o.f.WriteAt(p, off))
}

// WriteTo reads data from the file and writes it to w until EOF.
// It implements the io.WriterTo interface for efficient file copying.
// The EOF callback is invoked when the end of file is reached.
// Returns ErrorNilPointer if called on nil instance, nil writer, or closed file.
func (o *progress) WriteTo(w io.Writer) (n int64, err error) {
	if o == nil || w == nil || o.f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	var bf = make([]byte, o.getBufferSize(0))

	for {
		var (
			nr int
			nw int
			er error
			ew error
		)

		// code from io.copy

		nr, er = o.f.Read(bf)

		if nr > 0 {
			nw, ew = w.Write(bf[:nr])
		}

		n += int64(nw)

		if er != nil && errors.Is(er, io.EOF) {
			o.finish()
			break
		} else if er != nil {
			err = er
			break
		} else if ew != nil {
			err = ew
			break
		} else if nw < nr {
			err = io.ErrShortWrite
			break
		} else if nw != nr {
			err = errors.New("invalid write result")
			break
		}

		clear(bf)
	}

	return n, err
}

// WriteString writes the contents of string s to the file.
// It implements the io.StringWriter interface with integrated progress tracking.
// The increment callback is invoked with the number of bytes written.
// Returns ErrorNilPointer if called on nil instance or closed file.
func (o *progress) WriteString(s string) (n int, err error) {
	if o == nil || o.f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	return o.analyze(o.f.WriteString(s))
}
