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
	"os"
	"path/filepath"

	arctps "github.com/nabbar/golib/archive/archive/types"
)

// wrt implements the arctps.Writer interface for tar archives.
// It wraps an io.WriteCloser and provides methods to add files and directories
// to a tar format archive.
type wrt struct {
	w io.WriteCloser // The underlying writer where tar archive data is written
	z *tar.Writer    // The standard library tar writer for creating the archive
}

// Close finalizes the tar archive and closes the underlying writer.
// This method must be called to ensure the archive is properly terminated with
// end-of-archive markers. The operations are performed in sequence:
//  1. Flush any buffered data in the tar writer
//  2. Close the tar writer (writes end-of-archive markers)
//  3. Close the underlying io.WriteCloser
//
// Returns:
//   - error: The first error encountered during the close sequence, if any.
//
// Important: Failure to call Close() will result in a corrupted or incomplete archive.
// Always call Close() before closing or discarding the underlying writer.
//
// Example:
//
//	defer func() {
//	    if err := writer.Close(); err != nil {
//	        log.Printf("Error closing archive: %v", err)
//	    }
//	}()
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

// Add adds a file to the tar archive.
//
// It takes in the file information, the file reader, and the target path if the new file is a link.
// It returns an error if any operation fails.
func (o *wrt) Add(i fs.FileInfo, r io.ReadCloser, forcePath, target string) error {
	var (
		e error
		h *tar.Header
	)

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	if h, e = tar.FileInfoHeader(i, target); e != nil {
		return e
	} else if len(target) > 0 {
		h.Linkname = target
	}

	if len(forcePath) > 0 {
		h.Name = forcePath
	}

	if e = o.z.WriteHeader(h); e != nil {
		return e
	}

	if r != nil {
		if _, e = io.Copy(o.z, r); e != nil {
			return e
		}
	}

	return nil
}

// FromPath adds files from a source path to the tar archive, optionally filtering and renaming them.
// If the source is a single file, it is added directly. If the source is a directory,
// all files within are walked recursively and added to the archive.
//
// Parameters:
//   - source: The filesystem path to add. Can be a file or directory.
//   - filter: A glob pattern to filter files (e.g., "*.txt", "*.go"). Use "*" or empty string
//     to include all files. The pattern is matched against the full file path.
//   - fct: An optional function to transform file paths in the archive. If nil, the original
//     paths are used. This is useful for stripping prefixes or reorganizing the archive structure.
//
// Returns:
//   - error: Any error encountered during walking or adding files.
//
// Behavior:
//   - Regular files are added with their contents
//   - Directories are skipped (only their contents are added)
//   - Symbolic links are preserved with their target paths
//   - Hard links are treated as symbolic links
//   - Other file types (devices, pipes) return fs.ErrInvalid
//
// Example:
//
//	// Add all Go files from a directory, stripping the prefix
//	err := writer.FromPath("/home/user/project", "*.go", func(path string) string {
//	    return strings.TrimPrefix(path, "/home/user/project/")
//	})
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

// addFiltering is an internal helper that filters and adds a single file to the archive.
// This method implements the filtering logic and file type handling for FromPath.
//
// Parameters:
//   - source: The filesystem path of the file to add.
//   - filter: A glob pattern to match against the file path.
//   - fct: An optional function to transform the file path in the archive.
//   - info: The FileInfo for the file to add.
//
// Returns:
//   - error: fs.ErrInvalid for unsupported file types, or any I/O error encountered.
//
// This method handles:
//   - Applying the filter pattern (default "*" includes all files)
//   - Skipping directories (returns nil without adding)
//   - Reading symbolic and hard links to preserve their targets
//   - Opening regular files for content copying
//   - Rejecting unsupported file types (devices, pipes, etc.)
func (o *wrt) addFiltering(source string, filter string, fct arctps.ReplaceName, info fs.FileInfo) error {
	var (
		ok     bool
		err    error
		rpt    *os.Root
		hdf    *os.File
		target string
	)

	// Set default filter to match all files if not specified
	if len(filter) < 1 {
		filter = "*"
	}

	// Set default name replacement function (identity function) if not provided
	if fct == nil {
		fct = func(source string) string {
			return source
		}
	}

	// Apply filter pattern on filename - skip files that don't match
	if ok, err = filepath.Match(filter, filepath.Base(source)); err != nil {
		return err
	} else if !ok {
		return nil
	}

	// Validate file info and handle different file types
	if info == nil {
		return fs.ErrInvalid
	} else if info.IsDir() {
		// Skip directories - they are implicitly created by their contents
		return nil
	} else if info.Mode()&os.ModeSymlink != 0 {
		// Symbolic link - read the target path
		if target, err = os.Readlink(source); err != nil {
			return err
		}
	} else if info.Mode()&os.ModeDevice != 0 {
		// Hard link - read the target path
		if target, err = os.Readlink(source); err != nil {
			return err
		}
	} else if info.Mode().IsRegular() {
		// Regular file - open for reading
		rpt, err = os.OpenRoot(filepath.Dir(source))

		defer func() {
			if rpt != nil {
				_ = rpt.Close()
			}
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
		// Unsupported file type (device, pipe, socket, etc.)
		return fs.ErrInvalid
	}

	return o.Add(info, hdf, fct(source), target)
}
