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
	"path/filepath"

	arctps "github.com/nabbar/golib/archive/archive/types"
)

// wrt is the internal implementation of the types.Writer interface for ZIP archives.
// It wraps both the underlying io.WriteCloser and the archive/zip.Writer to provide
// unified archive creation.
type wrt struct {
	// w is the underlying io.WriteCloser where the ZIP archive data is written.
	// It is closed after the zip.Writer is closed.
	w io.WriteCloser
	// z is the zip.Writer that handles ZIP format creation and file compression.
	z *zip.Writer
}

// Close finalizes the ZIP archive and closes all underlying resources.
// It performs the following operations in sequence:
//  1. Flushes the zip.Writer to ensure all buffered data is written
//  2. Closes the zip.Writer to write the central directory
//  3. Closes the underlying io.WriteCloser
//
// If any step fails, the error from the first failure is returned and subsequent
// steps are skipped.
//
// Returns:
//   - error: The first error encountered during the close sequence, or nil if successful.
func (o *wrt) Close() error {
	if e := o.z.Flush(); e != nil {
		return e
	} else if e = o.z.Close(); e != nil {
		return e
	} else if e = o.w.Close(); e != nil {
		return e
	}

	return nil
}

// Add adds a single file to the ZIP archive with the given file information and content.
//
// The method automatically closes the provided io.ReadCloser via defer, so the caller
// should not close it. If the reader is nil, the method returns nil without error.
//
// Parameters:
//   - i: File information (fs.FileInfo) containing metadata like size, permissions, and timestamps.
//     This information is used to create the ZIP file header.
//   - r: An io.ReadCloser providing the file content to be added to the archive.
//     If nil, the method returns nil (no-op). The reader is automatically closed.
//   - forcePath: If non-empty, overrides the file name in the archive header. Use this to
//     store the file under a different path than indicated by fs.FileInfo.
//   - notUse: Currently unused parameter, kept for interface compatibility.
//
// Returns:
//   - error: Any error encountered during header creation, file creation in archive,
//     or content copying. Returns nil if r is nil.
func (o *wrt) Add(i fs.FileInfo, r io.ReadCloser, forcePath, notUse string) error {
	var (
		e error
		h *zip.FileHeader
		w io.Writer
	)

	if r == nil {
		return nil
	}

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	if h, e = zip.FileInfoHeader(i); e != nil {
		return e
	} else if len(forcePath) > 0 {
		h.Name = forcePath
	}

	if w, e = o.z.CreateHeader(h); e != nil {
		return e
	} else if _, e = io.Copy(w, r); e != nil {
		return e
	}

	return nil
}

// FromPath recursively adds files from a filesystem path to the ZIP archive.
//
// The method walks the directory tree starting from the source path and adds all
// regular files that match the filter pattern. Directories and non-regular files
// (symlinks, devices, etc.) are skipped.
//
// Parameters:
//   - source: The filesystem path to add. Can be a file or directory.
//     If a file, only that file is added (subject to filter).
//     If a directory, all matching files in the tree are added recursively.
//   - filter: A glob pattern for filtering files (e.g., "*.txt", "data_*.json").
//     Uses filepath.Match for pattern matching. If empty, defaults to "*" (all files).
//   - fct: Optional function to transform file paths before adding to archive.
//     If nil, files are added with their original source paths.
//     The function receives the source path and returns the desired archive path.
//
// Returns:
//   - error: Any error encountered during directory walking, file opening, or archive addition.
//     Returns fs.ErrInvalid for non-regular files or invalid file info.
func (o *wrt) FromPath(source string, filter string, fct arctps.ReplaceName) error {
	if i, e := os.Stat(source); e == nil && !i.IsDir() {
		return o.addFiltering(source, filter, fct, i)
	}

	return filepath.Walk(source, func(path string, info fs.FileInfo, e error) error {
		if e != nil {
			return e
		}

		return o.addFiltering(path, filter, fct, info)
	})
}

// addFiltering is an internal method that applies filtering and adds a single file to the archive.
//
// This method performs the following operations:
//  1. Applies the filter pattern to the source path
//  2. Skips directories (returns nil)
//  3. Opens regular files using os.OpenRoot for secure access
//  4. Applies the path replacement function if provided
//  5. Adds the file to the archive via Add method
//
// Parameters:
//   - source: The filesystem path of the file to potentially add.
//   - filter: Glob pattern for filtering. Defaults to "*" if empty.
//   - fct: Path transformation function. Defaults to identity function if nil.
//   - info: File metadata. Must not be nil.
//
// Returns:
//   - error: Returns nil if file doesn't match filter or is a directory.
//     Returns fs.ErrInvalid if info is nil or file is not regular.
//     Returns any error from pattern matching, file opening, or archive addition.
func (o *wrt) addFiltering(source string, filter string, fct arctps.ReplaceName, info fs.FileInfo) error {
	var (
		ok  bool
		err error
		rpt *os.Root
		hdf *os.File
	)

	if len(filter) < 1 {
		filter = "*"
	}

	if fct == nil {
		fct = func(source string) string {
			return source
		}
	}

	if ok, err = filepath.Match(filter, source); err != nil {
		return err
	} else if !ok {
		return nil
	}

	if info == nil {
		return fs.ErrInvalid
	} else if info.IsDir() {
		return nil
	} else if info.Mode().IsRegular() {
		rpt, err = os.OpenRoot(filepath.Dir(source))

		defer func() {
			_ = rpt.Close()
		}()

		if err != nil {
			return err
		}

		hdf, err = rpt.Open(filepath.Base(source))

		defer func() {
			_ = hdf.Close()
		}()

		if err != nil {
			return err
		}
	} else {
		return fs.ErrInvalid
	}

	return o.Add(info, hdf, fct(source), "")
}
