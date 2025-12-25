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

package tar

import (
	"archive/tar"
	"io"
	"io/fs"

	arctps "github.com/nabbar/golib/archive/archive/types"
)

// reset defines an interface for readers that support rewinding to the beginning.
// If the underlying io.ReadCloser implements this interface, the tar reader can
// reset its position to re-read the archive from the start.
type reset interface {
	// Reset attempts to reset the reader to its initial position.
	// Returns true if the reset was successful, false otherwise.
	Reset() bool
}

// rdr implements the arctps.Reader interface for tar archives.
// It wraps an io.ReadCloser and provides methods to read and extract files
// from tar format archives.
type rdr struct {
	r io.ReadCloser // The underlying reader containing tar archive data
	z *tar.Reader   // The standard library tar reader for parsing the archive
}

// Reset attempts to reset the reader to the beginning of the archive.
// This method checks if the underlying io.ReadCloser implements the reset interface.
// If supported, it resets the reader position and creates a new tar.Reader.
//
// Returns:
//   - bool: true if reset was successful, false if the underlying reader doesn't support reset.
//
// Note: Most file-based readers support reset via seeking. Network streams typically do not.
func (o *rdr) Reset() bool {
	if r, k := o.r.(reset); k {
		return r.Reset()
	}

	return false
}

// Close closes the underlying io.ReadCloser, releasing any associated resources.
//
// Returns:
//   - error: Any error returned by the underlying reader's Close method.
//
// After calling Close, the reader should not be used for further operations.
func (o *rdr) Close() error {
	return o.r.Close()
}

// List returns a slice containing the paths of all files in the tar archive.
// This method iterates through the entire archive to enumerate all entries.
//
// If the underlying reader supports reset, the reader is reset to the beginning
// before listing. Otherwise, the listing starts from the current position.
//
// Returns:
//   - []string: A slice containing the path of each file in the archive.
//   - error: Always returns nil. Errors during iteration are silently ignored.
//
// Note: After List() completes, the reader position is at the end of the archive.
// If you need to read files after listing, call Reset() or create a new reader.
//
// Example:
//
//	files, err := reader.List()
//	for _, path := range files {
//	    fmt.Println(path)
//	}
func (o *rdr) List() ([]string, error) {
	var (
		e error
		h *tar.Header
		l = make([]string, 0)
	)

	if o.Reset() {
		o.z = tar.NewReader(o.r)
	}

	for e == nil {
		h, e = o.z.Next()
		if h != nil {
			l = append(l, h.Name)
			_, _ = io.Copy(io.Discard, o.z)
		}
	}

	return l, nil
}

// Info retrieves file information for a specific file in the tar archive.
// This method searches for the file with the given path and returns its metadata.
//
// If the underlying reader supports reset, the reader is reset to the beginning
// before searching. Otherwise, the search starts from the current position.
//
// Parameters:
//   - s: The path of the file within the archive to get information for.
//
// Returns:
//   - fs.FileInfo: File metadata including name, size, permissions, and modification time.
//   - error: fs.ErrNotExist if the file is not found in the archive.
//
// Note: After Info() completes, the reader position may have advanced. If the file
// was not found, the position is at the end of the archive.
//
// Example:
//
//	info, err := reader.Info("path/to/file.txt")
//	if err == nil {
//	    fmt.Printf("Size: %d, Mode: %v\n", info.Size(), info.Mode())
//	}
func (o *rdr) Info(s string) (fs.FileInfo, error) {
	var (
		e error
		h *tar.Header
	)

	if o.Reset() {
		o.z = tar.NewReader(o.r)
	}

	for e == nil {
		h, e = o.z.Next()
		if h != nil && h.Name == s {
			return h.FileInfo(), nil
		} else if h != nil {
			_, _ = io.Copy(io.Discard, o.z)
		}
	}

	return nil, fs.ErrNotExist

}

// Get retrieves a specific file from the tar archive as an io.ReadCloser.
// This method searches for the file with the given path and returns a reader for its contents.
//
// If the underlying reader supports reset, the reader is reset to the beginning
// before searching. Otherwise, the search starts from the current position.
//
// Parameters:
//   - s: The path of the file within the archive to extract.
//
// Returns:
//   - io.ReadCloser: A reader for the file contents. This reader is a NopCloser wrapping
//     the tar reader, so closing it does not affect the underlying archive reader.
//   - error: fs.ErrNotExist if the file is not found in the archive.
//
// Important: The returned io.ReadCloser must be fully read before calling any other methods
// on the archive reader. Subsequent operations will advance the tar stream.
//
// Example:
//
//	rc, err := reader.Get("path/to/file.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer rc.Close()
//	data, _ := io.ReadAll(rc)
func (o *rdr) Get(s string) (io.ReadCloser, error) {
	var (
		e error
		h *tar.Header
	)

	if o.Reset() {
		o.z = tar.NewReader(o.r)
	}

	for e == nil {
		h, e = o.z.Next()
		if h != nil && h.Name == s {
			return io.NopCloser(o.z), nil
		} else if h != nil {
			_, _ = io.Copy(io.Discard, o.z)
		}
	}

	return nil, fs.ErrNotExist
}

// Has checks whether a file with the given path exists in the tar archive.
// This method searches through the archive to determine if the file is present.
//
// If the underlying reader supports reset, the reader is reset to the beginning
// before searching. Otherwise, the search starts from the current position.
//
// Parameters:
//   - s: The path of the file to check for within the archive.
//
// Returns:
//   - bool: true if the file exists in the archive, false otherwise.
//
// Note: This method may advance the reader position. If the file is not found,
// the reader position will be at the end of the archive.
//
// Example:
//
//	if reader.Has("path/to/file.txt") {
//	    fmt.Println("File exists in archive")
//	}
func (o *rdr) Has(s string) bool {
	var (
		e error
		h *tar.Header
	)

	if o.Reset() {
		o.z = tar.NewReader(o.r)
	}

	for e == nil {
		h, e = o.z.Next()
		if h != nil && h.Name == s {
			return true
		} else if h != nil {
			_, _ = io.Copy(io.Discard, o.z)
		}
	}

	return false
}

// Walk iterates through all files in the tar archive, calling the provided function for each entry.
// This method processes each file in the archive sequentially, allowing custom extraction logic.
//
// If the underlying reader supports reset, the reader is reset to the beginning
// before walking. Otherwise, the walk starts from the current position.
//
// Parameters:
//   - fct: A callback function invoked for each file in the archive. The function receives:
//   - fs.FileInfo: File metadata (name, size, permissions, etc.)
//   - io.ReadCloser: A reader for the file contents (NopCloser, closing has no effect)
//   - string: The path of the file within the archive
//   - string: The link target if the file is a symbolic or hard link, empty otherwise
//
// The callback function must return:
//   - true: Continue walking to the next file
//   - false: Stop walking immediately
//
// Important: The io.ReadCloser provided to the callback must be fully consumed (read or discarded)
// before the callback returns. The Walk method ensures any remaining data is discarded after
// each callback to maintain archive integrity.
//
// Example:
//
//	reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
//	    fmt.Printf("File: %s, Size: %d\n", path, info.Size())
//	    if strings.HasSuffix(path, ".txt") {
//	        data, _ := io.ReadAll(rc)
//	        processTextFile(data)
//	    }
//	    return true // Continue to next file
//	})
func (o *rdr) Walk(fct arctps.FuncExtract) {
	var (
		e error
		h *tar.Header
	)

	if o.Reset() {
		o.z = tar.NewReader(o.r)
	}

	for e == nil {
		h, e = o.z.Next()

		if h == nil || e != nil {
			continue
		}

		if !fct(h.FileInfo(), io.NopCloser(o.z), h.Name, h.Linkname) {
			return
		}

		// prevent file cursor not at EOF of current file
		_, _ = io.Copy(io.Discard, o.z)
	}
}
