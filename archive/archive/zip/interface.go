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

package zip

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"

	arctps "github.com/nabbar/golib/archive/archive/types"
)

// readerSize is an interface that provides the Size method.
// It is used to determine the size of a ZIP archive for random access.
type readerSize interface {
	Size() int64
}
type readerStat interface {
	Stat() (os.FileInfo, error)
}

// readerAt is a composite interface that combines io.ReadCloser and io.ReaderAt.
// It is required for reading ZIP archives which need random access capabilities.
type readerAt interface {
	io.ReadCloser
	io.ReaderAt
}

// NewReader creates a new ZIP archive Reader from the given io.ReadCloser.
//
// The provided io.ReadCloser must implement three additional interfaces for ZIP reading:
//   - readerSize: Provides Size() method returning archive size in bytes
//   - readerAt: Provides ReadAt() method for random access reading
//   - io.Seeker: Provides Seek() method for positioning within the archive
//
// Common types that satisfy these requirements include *os.File and *bytes.Reader.
//
// The function performs the following validations:
//  1. Checks that r implements readerSize interface
//  2. Checks that r implements readerAt interface (io.ReadCloser + io.ReaderAt)
//  3. Checks that r implements io.Seeker interface
//  4. Verifies that the archive size is greater than 0
//  5. Seeks to the beginning of the stream (position 0)
//  6. Creates the underlying zip.Reader
//
// Parameters:
//   - r: An io.ReadCloser that must also implement readerSize, io.ReaderAt, and io.Seeker.
//
// Returns:
//   - arctps.Reader: A Reader interface implementation for accessing the ZIP archive.
//   - error: Returns fs.ErrInvalid if required interfaces are not implemented or if size is invalid.
//     Returns seek or zip.NewReader errors if those operations fail.
//
// Example:
//
//	f, err := os.Open("archive.zip")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer f.Close()
//
//	reader, err := zip.NewReader(f)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
func NewReader(r io.ReadCloser) (arctps.Reader, error) {
	var siz int64
	if rs, ok := r.(readerSize); ok {
		siz = rs.Size()
	} else if ri, ok := r.(readerStat); ok {
		i, e := ri.Stat()
		if e == nil {
			siz = i.Size()
		}
	}

	if siz <= 0 {
		if rs, ok := r.(io.Seeker); !ok {
			return nil, fs.ErrInvalid
		} else {
			_, e := rs.Seek(0, io.SeekStart)
			if e != nil {
				return nil, e
			}
			n, e := rs.Seek(0, io.SeekEnd)
			if e != nil {
				return nil, e
			}
			if n <= 0 {
				return nil, fs.ErrInvalid
			}
			siz = n
		}
	}

	if ra, ok := r.(readerAt); !ok {
		return nil, fs.ErrInvalid
	} else if rs, o := r.(io.Seeker); !o {
		return nil, fs.ErrInvalid
	} else if _, e := rs.Seek(0, io.SeekStart); e != nil {
		return nil, e
	} else if z, err := zip.NewReader(ra, siz); err != nil {
		return nil, err
	} else {
		return &rdr{
			r: r,
			z: z,
		}, nil
	}
}

// NewWriter creates a new ZIP archive Writer from the given io.WriteCloser.
//
// The Writer allows adding files to a ZIP archive through the Add and FromPath methods.
// The archive is finalized when Close() is called, which flushes all data and writes
// the central directory.
//
// Parameters:
//   - w: An io.WriteCloser where the ZIP archive will be written. This can be a file,
//     buffer, or any other writable destination.
//
// Returns:
//   - arctps.Writer: A Writer interface implementation for creating the ZIP archive.
//   - error: Always returns nil in the current implementation, but the error return
//     is kept for interface consistency and future enhancements.
//
// Example:
//
//	f, err := os.Create("archive.zip")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer f.Close()
//
//	writer, err := zip.NewWriter(f)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer writer.Close()
func NewWriter(w io.WriteCloser) (arctps.Writer, error) {
	return &wrt{
		w: w,
		z: zip.NewWriter(w),
	}, nil
}
