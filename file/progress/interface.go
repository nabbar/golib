/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

// Package progress provides file I/O operations with real-time progress tracking and callbacks.
//
// This package wraps standard file operations with integrated progress monitoring capabilities,
// enabling applications to track and respond to file I/O events through registered callbacks.
// It implements all standard Go I/O interfaces while adding progress tracking functionality.
//
// Key features:
//   - Progress tracking with callbacks (increment, reset, EOF)
//   - Full io.Reader, io.Writer, io.Seeker interface implementation
//   - Configurable buffer sizes for optimal performance
//   - Temporary and unique file creation
//   - File position tracking (BOF/EOF)
//   - Thread-safe atomic operations
//
// Example usage:
//
//	import (
//	    "io"
//	    "github.com/nabbar/golib/file/progress"
//	)
//
//	// Open file with progress tracking
//	p, err := progress.Open("largefile.dat")
//	if err != nil {
//	    panic(err)
//	}
//	defer p.Close()
//
//	// Register progress callback
//	p.RegisterFctIncrement(func(bytes int64) {
//	    fmt.Printf("Progress: %d bytes\n", bytes)
//	})
//
//	// Register EOF callback
//	p.RegisterFctEOF(func() {
//	    fmt.Println("Reading complete!")
//	})
//
//	// Read file - callbacks will be invoked
//	io.Copy(io.Discard, p)
package progress

import (
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
)

const DefaultBuffSize = 32 * 1024 // see io.copyBuffer

type FctIncrement func(size int64)
type FctReset func(size, current int64)
type FctEOF func()

type GenericIO interface {
	io.ReadCloser
	io.ReadSeeker
	io.ReadWriteCloser
	io.ReadWriteSeeker
	io.WriteCloser
	io.WriteSeeker
	io.Reader
	io.ReaderFrom
	io.ReaderAt
	io.Writer
	io.WriterAt
	io.WriterTo
	io.Seeker
	io.StringWriter
	io.Closer
	io.ByteReader
	io.ByteWriter
}

type File interface {
	// CloseDelete closes and deletes the file if it is a regular file.
	//
	// It returns an error if the file is not a regular file, or if the file
	// cannot be closed or deleted.
	//
	// Note that this method does not follow symlinks. Therefore, if the file is a
	// symbolic link, the target of the symbolic link will not be deleted.
	// If you want to delete the target of the symbolic link, you should use the
	// os.Readlink function to get the target path and then the os.Remove function to
	// delete the target.
	CloseDelete() error

	// Path returns the path of the file.
	Path() string
	// Stat returns the file information as os.FileInfo for the given file or an error.
	//
	// If the file is not a regular file, or if the file information cannot be retrieved,
	// the function will return an error.
	//
	// Note that this method does not follow symlinks. Therefore, if the file is a symbolic link,
	// the target of the symbolic link will not be retrieved. If you want to retrieve the
	// target of the symbolic link, you should use the os.Readlink function to get the target path
	// and then the os.Stat function to retrieve the file information.
	Stat() (os.FileInfo, error)

	// SizeBOF returns the size of the file before the end of file (BOF) and an error if the size cannot be retrieved.
	//
	// The size is returned as an int64 and the error is returned as an error interface.
	//
	// If the file is not a regular file, or if the file information cannot be retrieved, the function will return an error.
	//
	// Note that this method does not follow symlinks. Therefore, if the file is a symbolic link, the target of the
	// symbolic link will not be retrieved. If you want to retrieve the target of the symbolic link, you should use
	// the os.Readlink function to get the target path and then the os.Stat function to retrieve the file information.
	SizeBOF() (size int64, err error)
	// SizeEOF returns the size of the file from the current position to the end of the file (EOF)
	// and an error if the size cannot be retrieved.
	//
	// The size is returned as an int64 and the error is returned as an error interface.
	//
	// If the file is not a regular file, or if the file information cannot be retrieved, the function will return an error.
	//
	// Note that this method does not follow symlinks. Therefore, if the file is a symbolic link, the target of the
	// symbolic link will not be retrieved. If you want to retrieve the target of the symbolic link, you should use
	// the os.Readlink function to get the target path and then the os.Stat function to retrieve the file information.
	SizeEOF() (size int64, err error)

	// Truncate will truncate the file at the given size.
	//
	// The function will return an error if the file cannot be truncated at the given size.
	//
	// Note that this method does not follow symlinks. Therefore, if the file is a symbolic link, the target of the
	// symbolic link will not be truncated. If you want to truncate the target of the symbolic link, you should use
	// the os.Readlink function to get the target path and then the os.Truncate function to truncate the target file.
	Truncate(size int64) error
	// Sync calls the Sync method on the underlying file.
	//
	// It returns an error if the Sync method cannot be called or if it returns an error.
	//
	// Note that this method does not follow symlinks. Therefore, if the file is a symbolic link, the target of the
	// symbolic link will not be synced. If you want to sync the target of the symbolic link, you should use
	// the os.Readlink function to get the target path and then the os.File.Sync function to sync the target file.
	Sync() error
}

type TempFile interface {
	IsTemp() bool
}

type Progress interface {
	GenericIO
	File
	TempFile

	// RegisterFctIncrement registers a function to be called when the progress of
	// a file being read or written reaches a certain number of bytes. The
	// function will be called with the number of bytes that have been read or
	// written from the start of the file. The function is called even if the
	// registered progress is not reached (i.e. if the file is smaller than
	// the registered progress). The function is called with the current
	// progress when the file is closed (i.e. when io.Copy returns io.EOF).
	//
	// The function is called with the following signature:
	//
	// func(size int64)
	//
	// If the function is nil, it is simply ignored.
	RegisterFctIncrement(fct FctIncrement)
	// RegisterFctReset registers a function to be called when the progress of a
	// file being read or written is reset. The function will be called with the
	// maximum progress that has been reached and the current progress when
	// the file is closed (i.e. when io.Copy returns io.EOF).
	//
	// The function is called with the following signature:
	//
	// func(size, current int64)
	//
	// If the function is nil, it is simply ignored.
	RegisterFctReset(fct FctReset)
	// RegisterFctEOF registers a function to be called when the end of a file is reached.
	// The function will be called with no arguments and will be called even if the
	// registered progress is not reached (i.e. if the file is smaller than
	// the registered progress).
	//
	// If the function is nil, it is simply ignored.
	RegisterFctEOF(fct FctEOF)
	// SetBufferSize sets the buffer size for the progress.
	//
	// The buffer size is used for the io.Copy function when reading from or writing to a file.
	//
	// The buffer size should be a power of two and greater or equal to 1.
	//
	// If the buffer size is less than 1 or if it is not a power of two, the function will return an error.
	//
	// The function returns an error if the buffer size cannot be set.
	//
	// Note that this method does not follow symlinks. Therefore, if the file is a symbolic link, the target of the
	// symbolic link will not be affected. If you want to set the buffer size of the target of the symbolic link, you
	// should use the os.Readlink function to get the target path and then the os.File.SetBufferSize function to set the
	// buffer size of the target file.
	SetBufferSize(size int32)
	// SetRegisterProgress sets the progress of the file.
	//
	// It takes a Progress interface as argument and sets the progress of the file to the given progress.
	//
	// The progress is used when the io.Copy function is called to read from or write to the file.
	//
	// The progress is used to call the functions registered with RegisterFctIncrement, RegisterFctReset and RegisterFctEOF.
	//
	// If the progress is nil, it is simply ignored.
	//
	// The function returns an error if the progress cannot be set.
	SetRegisterProgress(f Progress)

	// Reset resets the progress of the file to the given maximum progress.
	//
	// It is used to reset the progress of the file after it has been closed.
	//
	// If the maximum progress is less than 1, the function will return an error.
	//
	// The function returns an error if the progress cannot be reset.
	Reset(max int64)
}

// New opens a file with the given name, flags and permissions and returns a Progress interface.
//
// The Progress interface is used to track the progress of the file when it is being read from or written to.
//
// The function returns an error if the file cannot be opened.
func New(name string, flags int, perm os.FileMode) (Progress, error) {
	r, e := os.OpenRoot(filepath.Dir(name))
	if e != nil {
		return nil, e
	}

	f, e := r.OpenFile(filepath.Base(name), flags, perm)
	if e != nil {
		return nil, e
	}

	return &progress{
		r:  r,
		f:  f,
		t:  false,
		b:  new(atomic.Int32),
		fi: new(atomic.Value),
		fe: new(atomic.Value),
		fr: new(atomic.Value),
	}, nil
}

// Unique creates a new unique file in the given base path with the given pattern.
//
// The pattern should follow the rules of the os.CreateTemp function.
//
// The function returns a Progress interface and an error. The Progress interface is used to track the
// progress of the file when it is being read from or written to.
//
// The function returns an error if the file cannot be created.
func Unique(basePath, pattern string) (Progress, error) {
	f, e := os.CreateTemp(basePath, pattern)

	if e != nil {
		return nil, e
	}

	return &progress{
		r:  nil,
		f:  f,
		t:  false,
		b:  new(atomic.Int32),
		fi: new(atomic.Value),
		fe: new(atomic.Value),
		fr: new(atomic.Value),
	}, nil
}

// Temp creates a new temporary file with the given pattern and returns a Progress interface.
//
// The pattern should follow the rules of the os.CreateTemp function.
//
// The function returns an error if the file cannot be created.
//
// The Progress interface is used to track the progress of the file when it is being read from or written to.
//
// The function returns an error if the progress cannot be set.
//
// The function is used to create temporary files that are automatically deleted when they are closed.
func Temp(pattern string) (Progress, error) {
	f, e := os.CreateTemp("", pattern)

	if e != nil {
		return nil, e
	}

	return &progress{
		r:  nil,
		f:  f,
		t:  true,
		b:  new(atomic.Int32),
		fi: new(atomic.Value),
		fe: new(atomic.Value),
		fr: new(atomic.Value),
	}, nil
}

// Open opens the file with the given name and returns a Progress interface.
//
// The function returns an error if the file cannot be opened.
//
// The Progress interface is used to track the progress of the file when it is being read from or written to.
//
// The function returns an error if the progress cannot be set.
//
// The function is used to open existing files. If the file does not exist, the function will return an error.
func Open(name string) (Progress, error) {
	r, e := os.OpenRoot(filepath.Dir(name))
	if e != nil {
		return nil, e
	}

	f, e := r.Open(filepath.Base(name))
	if e != nil {
		return nil, e
	}

	return &progress{
		r:  r,
		f:  f,
		t:  false,
		b:  new(atomic.Int32),
		fi: new(atomic.Value),
		fe: new(atomic.Value),
		fr: new(atomic.Value),
	}, nil
}

// Create creates a new file with the given name and returns a Progress interface.
//
// The function returns an error if the file cannot be created.
//
// The Progress interface is used to track the progress of the file when it is being read from or written to.
//
// The function returns an error if the progress cannot be set.
//
// The function is used to create new files. If the file already exists, the function will return an error.
func Create(name string) (Progress, error) {
	r, e := os.OpenRoot(filepath.Dir(name))
	if e != nil {
		return nil, e
	}

	f, e := r.Create(filepath.Base(name))
	if e != nil {
		return nil, e
	}

	return &progress{
		r:  r,
		f:  f,
		t:  false,
		b:  new(atomic.Int32),
		fi: new(atomic.Value),
		fe: new(atomic.Value),
		fr: new(atomic.Value),
	}, nil
}
