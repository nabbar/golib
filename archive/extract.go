/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package archive

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	arcarc "github.com/nabbar/golib/archive/archive"
	arctps "github.com/nabbar/golib/archive/archive/types"
	arccmp "github.com/nabbar/golib/archive/compress"
)

// ExtractAll automatically detects and extracts archive and/or compressed files.
// It handles nested compression (e.g., .tar.gz) by recursively detecting formats.
//
// Parameters:
//   - r: The input reader containing the archive/compressed data
//   - archiveName: Original filename (used to strip compression extensions)
//   - destination: Target directory for extracted files
//
// Returns an error if extraction fails or if the input is invalid.
func ExtractAll(r io.ReadCloser, archiveName, destination string) error {
	var (
		e error
		n string
		a arccmp.Algorithm
		o io.ReadCloser
	)

	if r == nil {
		return fs.ErrInvalid
	}

	// Step 1: Detect and handle compression layer
	a, o, e = DetectCompression(r)

	// If compression is detected, decompress and recurse to handle underlying archive
	// Example: file.tar.gz -> decompress gzip -> recurse for tar
	if e == nil && !a.IsNone() && o != nil {
		n = strings.TrimSuffix(filepath.Base(archiveName), a.Extension())
		return ExtractAll(o, n, destination)
	}

	var (
		b arcarc.Algorithm
		z arctps.Reader
	)

	// Use original reader if no compression was detected to preserve ReaderAt interface for ZIP.
	// ZIP archives require random access (io.ReaderAt), which would be lost through decompression.
	// Reset read position since DetectCompression consumed header bytes.
	if a.IsNone() {
		if seeker, ok := r.(io.Seeker); ok {
			_, _ = seeker.Seek(0, io.SeekStart)
		}
		o = r
	}

	// Step 2: Detect and handle archive format (TAR, ZIP, or uncompressed file)
	if b, z, r, e = DetectArchive(o); e != nil {
		return e
	} else if b.IsNone() {
		// No archive format detected, treat as single file
		return writeFile(archiveName, destination, r, nil)
	} else if z == nil {
		return fs.ErrInvalid
	} else {
		var err error

		// Walk through all files in the archive
		z.Walk(func(info fs.FileInfo, closer io.ReadCloser, dst, target string) bool {
			defer func() {
				if closer != nil {
					_ = closer.Close()
				}
			}()

			// Handle different file types
			if info.IsDir() {
				// Create directory with preserved permissions
				if e = createPath(filepath.Join(destination, cleanPath(dst)), info.Mode()); e != nil {
					err = e
					return false
				}
			} else if info.Mode()&os.ModeSymlink != 0 {
				// Create symbolic link
				if e = writeSymLink(true, dst, target, destination); e != nil {
					err = e
					return false
				}
			} else if info.Mode()&os.ModeDevice != 0 {
				// Create hard link
				if e = writeSymLink(false, dst, target, destination); e != nil {
					err = e
					return false
				}
			} else if info.Mode().IsRegular() {
				// Extract regular file
				if e = writeFile(dst, destination, closer, info); e != nil {
					err = e
					return false
				}
			}

			// For TAR archives: ensure file cursor is at EOF before moving to next file
			// This prevents corruption when the caller doesn't fully read the file
			_, _ = io.Copy(io.Discard, closer)
			return true
		})

		return err
	}
}

// cleanPath removes directory traversal attempts (../) from paths to prevent
// security vulnerabilities where malicious archives could write outside the
// extraction directory.
func cleanPath(path string) string {
	for strings.Contains(path, ".."+string(filepath.Separator)) {
		path = filepath.Clean(path)
	}

	return path
}

// createPath recursively creates directories with the specified permissions.
// If the directory already exists, it validates that it's actually a directory.
// Uses default permissions (0750) if info is 0.
func createPath(dest string, info os.FileMode) error {
	if i, err := os.Stat(dest); err == nil {
		if i.IsDir() {
			return nil
		} else {
			return os.ErrInvalid
		}
	} else if os.IsNotExist(err) {
		// Recursively create parent directories
		if err = createPath(filepath.Dir(dest), info); err != nil {
			return err
		} else if info != 0 {
			return os.Mkdir(dest, info)
		} else {
			return os.Mkdir(dest, 0750)
		}
	} else {
		return err
	}
}

// writeFile extracts a single file from the archive to the destination directory.
// It sanitizes the path, creates parent directories, and preserves file permissions
// if provided in the FileInfo.
func writeFile(name, dest string, r io.ReadCloser, i fs.FileInfo) error {
	var (
		dst = filepath.Join(dest, cleanPath(name))
		hdf *os.File
		rpt *os.Root
		err error
	)

	defer func() {
		if hdf != nil {
			_ = hdf.Sync()
			_ = hdf.Close()
		}
		if rpt != nil {
			_ = rpt.Close()
		}
	}()

	// Create parent directories and open root for secure file creation
	if err = createPath(filepath.Dir(dst), 0); err != nil {
		return err
	} else if rpt, err = os.OpenRoot(filepath.Dir(dst)); err != nil {
		return err
	} else if hdf, err = rpt.Create(filepath.Base(dst)); err != nil {
		return err
	} else if _, err = io.Copy(hdf, r); err != nil {
		return err
	} else if i != nil {
		// Preserve original file permissions
		if err = os.Chmod(dst, i.Mode()); err != nil {
			return err
		}
	}

	return nil
}

// writeSymLink creates a symbolic link or hard link at the destination.
// The isSymLink parameter determines the link type:
//   - true: creates a symbolic link (soft link)
//   - false: creates a hard link
func writeSymLink(isSymLink bool, name, target, dest string) error {
	var (
		dst = filepath.Join(dest, cleanPath(name))
		err error
	)

	if err = createPath(filepath.Dir(dst), 0); err != nil {
		return err
	} else if isSymLink {
		return os.Symlink(target, dst)
	} else {
		return os.Link(target, dst)
	}
}
