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

	for e == nil {
		a, o, e = DetectCompression(r)

		if a.IsNone() {
			break
		}

		n = strings.TrimSuffix(filepath.Base(archiveName), a.Extension())
		return ExtractAll(o, n, destination)
	}

	var (
		b arcarc.Algorithm
		z arctps.Reader
	)

	if b, z, r, e = DetectArchive(o); e != nil {
		return e
	} else if b.IsNone() {
		return writeFile(archiveName, destination, r, nil)
	} else if z == nil {
		return fs.ErrInvalid
	} else {
		var err error

		z.Walk(func(info fs.FileInfo, closer io.ReadCloser, dst, target string) bool {
			defer func() {
				if closer != nil {
					_ = closer.Close()
				}
			}()

			if info.IsDir() {
				if e = createPath(filepath.Join(destination, cleanPath(dst)), info.Mode()); e != nil {
					err = e
					return false
				}
			} else if info.Mode()&os.ModeSymlink != 0 {
				if e = writeSymLink(true, dst, target, destination); e != nil {
					err = e
					return false
				}
			} else if info.Mode()&os.ModeDevice != 0 {
				if e = writeSymLink(false, dst, target, destination); e != nil {
					err = e
					return false
				}
			} else if info.Mode().IsRegular() {
				if e = writeFile(dst, destination, closer, info); e != nil {
					err = e
					return false
				}
			}

			// prevent file cursor not at EOF of current file for TAPE Archive
			_, _ = io.Copy(io.Discard, closer)
			return true
		})

		return err
	}
}

func cleanPath(path string) string {
	for strings.Contains(path, ".."+string(filepath.Separator)) {
		path = filepath.Clean(path)
	}

	return path
}

func createPath(dest string, info os.FileMode) error {
	if i, err := os.Stat(dest); err == nil {
		if i.IsDir() {
			return nil
		} else {
			return os.ErrInvalid
		}
	} else if os.IsNotExist(err) {
		if err = createPath(filepath.Dir(dest), info); err != nil {
			return err
		} else if info != 0 {
			return os.Mkdir(dest, info)
		} else {
			return os.Mkdir(dest, 0755)
		}
	} else {
		return err
	}
}

func writeFile(name, dest string, r io.ReadCloser, i fs.FileInfo) error {
	var (
		dst = filepath.Join(dest, cleanPath(name))
		hdf *os.File
		err error
	)

	defer func() {
		if hdf != nil {
			_ = hdf.Sync()
			_ = hdf.Close()
		}
	}()

	if err = createPath(filepath.Dir(dst), 0); err != nil {
		return err
	} else if hdf, err = os.Create(dst); err != nil {
		return err
	} else if _, err = io.Copy(hdf, r); err != nil {
		return err
	} else if i != nil {
		if err = os.Chmod(dst, i.Mode()); err != nil {
			return err
		}
	}

	return nil
}

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
