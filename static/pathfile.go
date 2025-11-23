/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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
 */

package static

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	libfpg "github.com/nabbar/golib/file/progress"
	libbuf "github.com/nabbar/golib/ioutils/bufferReadCloser"
)

func (s *staticHandler) getSize() int64 {
	return s.siz.Load()
}

func (s *staticHandler) setSize(size int64) {
	s.siz.Store(size)
}

func (s *staticHandler) getBase() []string {
	i := s.bph.Load()
	if i == nil {
		return make([]string, 0)
	} else {
		return i
	}
}

func (s *staticHandler) setBase(base ...string) {
	if len(base) > 0 {
		s.bph.Store(base)
	} else {
		s.bph.Store(make([]string, 0))
	}
}

func (s *staticHandler) dirEntries(root string) ([]fs.DirEntry, error) {
	if root == "" {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("pathfile is empty"))
	}

	val, err := s.efs.ReadDir(root)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.Error(err)
	} else if err != nil {
		return nil, ErrorFileOpen.Error(err)
	} else {
		return val, nil
	}
}

func (s *staticHandler) fileGet(pathFile string) (fs.FileInfo, io.ReadCloser, error) {
	if len(pathFile) < 1 {
		return nil, nil, ErrorParamEmpty.Error(fmt.Errorf("pathfile is empty"))
	}

	if inf, err := s.fileInfo(pathFile); err != nil {
		return nil, nil, err
	} else if inf.IsDir() {
		return inf, nil, ErrorFileNotFound.Error(fmt.Errorf("path is a directory: %spc", pathFile))
	} else if inf.Size() >= s.getSize() {
		r, e := s.fileTemp(pathFile)
		return inf, r, e
	} else {
		r, e := s.fileBuff(pathFile)
		return inf, r, e
	}
}

func (s *staticHandler) fileInfo(pathFile string) (fs.FileInfo, error) {
	if pathFile == "" {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("pathfile is empty"))
	}

	var inf fs.FileInfo
	obj, err := s.efs.Open(pathFile)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.Error(err)
	} else if err != nil {
		return nil, ErrorFileOpen.Error(err)
	}

	defer func() {
		_ = obj.Close()
	}()

	inf, err = obj.Stat()

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.Error(err)
	} else if err != nil {
		return nil, ErrorFileOpen.Error(err)
	}

	return inf, nil
}

func (s *staticHandler) fileBuff(pathFile string) (io.ReadCloser, error) {
	if pathFile == "" {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("pathfile is empty"))
	}

	obj, err := s.efs.ReadFile(pathFile)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.Error(err)
	} else if err != nil {
		return nil, ErrorFileOpen.Error(err)
	} else {
		return libbuf.NewBuffer(bytes.NewBuffer(obj), nil), nil
	}
}

func (s *staticHandler) fileTemp(pathFile string) (libfpg.Progress, error) {
	if pathFile == "" {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("pathfile is empty"))
	}

	var tmp libfpg.Progress
	obj, err := s.efs.Open(pathFile)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.Error(err)
	} else if err != nil {
		return nil, ErrorFileOpen.Error(err)
	}

	defer func() {
		_ = obj.Close()
	}()

	tmp, err = libfpg.Temp("")
	if err != nil {
		return nil, ErrorFiletemp.Error(err)
	}

	_, e := io.Copy(tmp, obj)
	if e != nil {
		return nil, ErrorFiletemp.Error(e)
	}

	// Reset cursor to beginning of file
	if _, e = tmp.Seek(0, io.SeekStart); e != nil {
		return nil, ErrorFiletemp.Error(e)
	}

	return tmp, nil
}

// Has checks if a file exists in the embedded filesystem.
func (s *staticHandler) Has(pathFile string) bool {
	if _, e := s.fileInfo(pathFile); e != nil {
		return false
	} else {
		return true
	}
}

// List returns all file paths under a root directory.
// Recursively walks the directory tree and returns relative paths.
func (s *staticHandler) List(rootPath string) ([]string, error) {
	var (
		err error
		res = make([]string, 0)
		lst []string
		ent []fs.DirEntry
		inf fs.FileInfo
	)

	if rootPath == "" {
		for _, p := range s.getBase() {
			inf, err = s.fileInfo(p)
			if err != nil {
				return nil, err
			}

			if !inf.IsDir() {
				res = append(res, p)
				continue
			}

			lst, err = s.List(p)

			if err != nil {
				return nil, err
			}

			res = append(res, lst...)
		}
	} else if ent, err = s.dirEntries(rootPath); err != nil {
		return nil, err
	} else {
		for _, f := range ent {

			if !f.IsDir() {
				res = append(res, path.Join(rootPath, f.Name()))
				continue
			}

			lst, err = s.List(path.Join(rootPath, f.Name()))

			if err != nil {
				return nil, err
			}

			res = append(res, lst...)
		}
	}

	return res, nil
}

// Find opens a file and returns a ReadCloser.
// The caller is responsible for closing the returned ReadCloser.
func (s *staticHandler) Find(pathFile string) (io.ReadCloser, error) {
	_, r, e := s.fileGet(pathFile)
	return r, e
}

// Info returns file information (size, modification time, etc.).
func (s *staticHandler) Info(pathFile string) (os.FileInfo, error) {
	return s.fileInfo(pathFile)
}

// Temp creates a temporary file copy with progress tracking.
// Useful for large files or when progress reporting is needed.
// See github.com/nabbar/golib/file/progress for details.
func (s *staticHandler) Temp(pathFile string) (libfpg.Progress, error) {
	return s.fileTemp(pathFile)
}

// Map iterates over all files in the embedded filesystem.
// The provided function is called for each file with its path and FileInfo.
// If the function returns an error, iteration stops and the error is returned.
func (s *staticHandler) Map(fct func(pathFile string, inf os.FileInfo) error) error {
	var (
		err error
		lst []string
		inf fs.FileInfo
	)

	if lst, err = s.List(""); err != nil {
		return err
	} else {
		for _, f := range lst {
			if inf, err = s.fileInfo(f); err != nil {
				return err
			} else if err = fct(f, inf); err != nil {
				return err
			}
		}
	}

	return nil
}

// UseTempForFileSize sets the threshold for using temporary files.
// Files larger than this size will be served via the Temp() method.
func (s *staticHandler) UseTempForFileSize(size int64) {
	s.setSize(size)
}
