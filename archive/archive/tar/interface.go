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

	arctps "github.com/nabbar/golib/archive/archive/types"
)

// NewReader creates a new tar archive reader from the provided io.ReadCloser.
//
// This function initializes a reader that can extract files from tar archives.
// The reader implements the arctps.Reader interface, providing methods to list,
// query, and extract files from the tar archive.
//
// Parameters:
//   - r: An io.ReadCloser containing the tar archive data. The caller is responsible
//     for closing this reader when done to release resources.
//
// Returns:
//   - arctps.Reader: A reader instance that implements the archive Reader interface.
//   - error: Always returns nil. This function does not fail during initialization.
//
// The returned reader supports:
//   - List(): Enumerate all files in the archive
//   - Info(path): Get file information for a specific path
//   - Get(path): Extract a specific file as io.ReadCloser
//   - Has(path): Check if a file exists in the archive
//   - Walk(func): Iterate over all files with a callback
//   - Reset(): Reset the reader to the beginning if the underlying reader supports it
//
// Important: The underlying io.ReadCloser must be closed by the caller to prevent
// resource leaks. The tar.Reader does not take ownership of the provided reader.
//
// Example:
//
//	file, err := os.Open("archive.tar")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	reader, err := tar.NewReader(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
//
//	// Use reader to access archive contents
//	files, err := reader.List()
func NewReader(r io.ReadCloser) (arctps.Reader, error) {
	return &rdr{
		r: r,
		z: tar.NewReader(r),
	}, nil
}

// NewWriter creates a new tar archive writer from the provided io.WriteCloser.
//
// This function initializes a writer that can create tar archives by adding files
// and directories. The writer implements the arctps.Writer interface, providing
// methods to add individual files or entire directory trees to the archive.
//
// Parameters:
//   - w: An io.WriteCloser where the tar archive data will be written. The caller
//     is responsible for closing this writer when done to ensure all data is flushed.
//
// Returns:
//   - arctps.Writer: A writer instance that implements the archive Writer interface.
//   - error: Always returns nil. This function does not fail during initialization.
//
// The returned writer supports:
//   - Add(info, reader, forcePath, target): Add a single file to the archive
//   - FromPath(source, filter, replaceName): Add files recursively from a directory
//   - Close(): Flush and close the archive (must be called before closing underlying writer)
//
// Important: The writer's Close() method must be called before closing the underlying
// io.WriteCloser to ensure the tar archive is properly finalized with end-of-archive
// markers. Failure to call Close() will result in a corrupted archive.
//
// Example:
//
//	file, err := os.Create("archive.tar")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	writer, err := tar.NewWriter(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer writer.Close() // Must close writer before file
//
//	// Add files to archive
//	err = writer.FromPath("/path/to/files", "*", nil)
func NewWriter(w io.WriteCloser) (arctps.Writer, error) {
	return &wrt{
		w: w,
		z: tar.NewWriter(w),
	}, nil
}
