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

	arctps "github.com/nabbar/golib/archive/archive/types"
)

// rdr is the internal implementation of the types.Reader interface for ZIP archives.
// It wraps both the underlying io.ReadCloser and the archive/zip.Reader to provide
// unified archive access.
type rdr struct {
	// r is the underlying io.ReadCloser that provides the raw ZIP archive data.
	// It is closed when Close() is called on the reader.
	r io.ReadCloser
	// z is the zip.Reader that handles ZIP format parsing and file extraction.
	z *zip.Reader
}

// Close closes the underlying io.ReadCloser, releasing any associated resources.
// After Close is called, no further operations should be performed on the reader.
//
// Returns:
//   - error: Any error encountered while closing the underlying reader.
func (o *rdr) Close() error {
	return o.r.Close()
}

// List returns a slice containing the names of all files in the ZIP archive.
// The returned slice is pre-allocated to the exact number of files for efficiency.
//
// Returns:
//   - []string: A slice of file names (paths) within the archive.
//   - error: Always returns nil in the current implementation.
func (o *rdr) List() ([]string, error) {
	var res = make([]string, 0, len(o.z.File))

	for _, f := range o.z.File {
		res = append(res, f.Name)
	}

	return res, nil
}

// Info returns the fs.FileInfo for the specified file in the ZIP archive.
// It performs a linear search through all files to find a matching name.
//
// Parameters:
//   - s: The path/name of the file within the archive to get information for.
//
// Returns:
//   - fs.FileInfo: File information including size, mode, modification time, etc.
//   - error: Returns fs.ErrNotExist if the specified file is not found in the archive.
func (o *rdr) Info(s string) (fs.FileInfo, error) {
	for _, f := range o.z.File {
		if f.Name == s {
			return f.FileInfo(), nil
		}
	}

	return nil, fs.ErrNotExist
}

// Get retrieves an io.ReadCloser for reading the contents of the specified file
// from the ZIP archive. The caller is responsible for closing the returned ReadCloser.
//
// Parameters:
//   - s: The path/name of the file within the archive to retrieve.
//
// Returns:
//   - io.ReadCloser: A reader for the file's decompressed content. Must be closed by caller.
//   - error: Returns fs.ErrNotExist if the file is not found, or any error from opening the file.
func (o *rdr) Get(s string) (io.ReadCloser, error) {
	for _, f := range o.z.File {
		if f.Name == s {
			return f.Open()
		}
	}

	return nil, fs.ErrNotExist
}

// Has checks whether the ZIP archive contains a file with the specified name.
// It performs a linear search through all files.
//
// Parameters:
//   - s: The path/name of the file to check for.
//
// Returns:
//   - bool: true if the file exists in the archive, false otherwise.
func (o *rdr) Has(s string) bool {
	for _, f := range o.z.File {
		if f.Name == s {
			return true
		}
	}

	return false
}

// Walk iterates through all files in the ZIP archive and calls the provided function
// for each file. The iteration stops if the callback function returns false.
//
// For each file, Walk opens the file and passes its information to the callback.
// The callback receives:
//   - fs.FileInfo: File metadata (size, permissions, timestamps)
//   - io.ReadCloser: Reader for file content (may be nil if open fails)
//   - string: File name/path within the archive
//   - string: Link target (always empty for ZIP archives)
//
// Note: Walk does not propagate errors from opening files. If a file cannot be opened,
// the callback is still called with a nil reader.
//
// Parameters:
//   - fct: Callback function called for each file. Return false to stop iteration.
func (o *rdr) Walk(fct arctps.FuncExtract) {
	for _, f := range o.z.File {
		r, _ := f.Open()
		if !fct(f.FileInfo(), r, f.Name, "") {
			return
		}
	}
}
