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
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
)

// progress implements the Progress interface with thread-safe progress tracking.
// It wraps os.File operations and provides callbacks for monitoring I/O operations.
type progress struct {
	r *os.Root // os Root for file operations
	f *os.File // underlying file handle
	t bool     // indicates if file is temporary (auto-deleted on close)

	b *atomic.Int32 // buffer size for I/O operations (atomic for thread-safety)

	fi *atomic.Value // increment callback function (FctIncrement)
	fe *atomic.Value // EOF callback function (FctEOF)
	fr *atomic.Value // reset callback function (FctReset)
}

// SetBufferSize sets the buffer size for I/O operations.
// The size is stored atomically to allow safe concurrent access.
// A size less than 1024 will result in using DefaultBuffSize.
func (o *progress) SetBufferSize(size int32) {
	o.b.Store(size)
}

// getBufferSize returns the buffer size to use for I/O operations.
// It prioritizes: 1) provided size parameter, 2) stored buffer size, 3) DefaultBuffSize.
// Minimum buffer size is 1024 bytes to ensure reasonable performance.
func (o *progress) getBufferSize(size int) int {
	if size > 0 {
		return size
	} else if o == nil {
		return DefaultBuffSize
	}

	i := o.b.Load()
	if i < 1024 {
		return DefaultBuffSize
	} else {
		return int(i)
	}
}

// IsTemp returns true if the file is a temporary file that will be automatically
// deleted when closed. Temporary files are created using Temp() or Unique() with
// auto-delete enabled.
func (o *progress) IsTemp() bool {
	return o.t
}

// Path returns the cleaned absolute path of the file.
// The path is cleaned using filepath.Clean to ensure canonical form.
func (o *progress) Path() string {
	return filepath.Clean(o.f.Name())
}

// Stat returns file information (os.FileInfo) for the underlying file.
// It wraps os.File.Stat() with proper error handling and nil checks.
// Returns ErrorNilPointer if called on nil instance or closed file.
// Returns ErrorIOFileStat if the stat operation fails.
func (o *progress) Stat() (os.FileInfo, error) {
	if o == nil || o.f == nil {
		return nil, ErrorNilPointer.Error(nil)
	}

	if i, e := o.f.Stat(); e != nil {
		return i, ErrorIOFileStat.Error(e)
	} else {
		return i, nil
	}
}

// SizeBOF returns the number of bytes from the beginning of the file (BOF)
// to the current position. This represents how many bytes have been read or written
// from the start of the file.
// Returns ErrorNilPointer if called on nil instance or closed file.
func (o *progress) SizeBOF() (size int64, err error) {
	if o == nil || o.f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	return o.seek(0, io.SeekCurrent)
}

// SizeEOF returns the number of bytes from the current position to the end of the file (EOF).
// This represents how many bytes remain to be read from the current position.
// The function preserves the current file position by seeking to EOF and back.
// Returns ErrorNilPointer if called on nil instance or closed file.
func (o *progress) SizeEOF() (size int64, err error) {
	if o == nil || o.f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	var (
		e error
		a int64 // origin position
		b int64 // eof position
	)

	if a, e = o.seek(0, io.SeekCurrent); e != nil {
		return 0, e
	} else if b, e = o.seek(0, io.SeekEnd); e != nil {
		return 0, e
	} else if _, e = o.seek(a, io.SeekStart); e != nil {
		return 0, e
	} else {
		return b - a, nil
	}
}

// Truncate changes the size of the file to the specified size.
// It wraps os.File.Truncate() and triggers the reset callback after truncation.
// Returns ErrorNilPointer if called on nil instance or closed file.
func (o *progress) Truncate(size int64) error {
	if o == nil || o.f == nil {
		return ErrorNilPointer.Error(nil)
	}

	e := o.f.Truncate(size)
	o.reset()

	return e
}

// Sync commits the current contents of the file to stable storage.
// It wraps os.File.Sync() with proper nil checks.
// Returns ErrorNilPointer if called on nil instance or closed file.
func (o *progress) Sync() error {
	if o == nil || o.f == nil {
		return ErrorNilPointer.Error(nil)
	}

	return o.f.Sync()
}
