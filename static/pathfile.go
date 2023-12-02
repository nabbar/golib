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

	liberr "github.com/nabbar/golib/errors"
	libfpg "github.com/nabbar/golib/file/progress"
	libbuf "github.com/nabbar/golib/ioutils/bufferReadCloser"
)

func (s *staticHandler) _getSize() int64 {
	return s.z.Load()
}

func (s *staticHandler) _setSize(size int64) {
	s.z.Store(size)
}

func (s *staticHandler) _getBase() []string {
	i := s.b.Load()
	if i == nil {
		return make([]string, 0)
	} else if b, k := i.([]string); !k {
		return make([]string, 0)
	} else {
		return b
	}
}

func (s *staticHandler) _setBase(base ...string) {
	var b = make([]string, 0)

	if len(base) > 0 {
		b = append(b, base...)
	}

	s.b.Store(b)
}

func (s *staticHandler) _listEmbed(root string) ([]fs.DirEntry, liberr.Error) {
	if root == "" {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("pathfile is empty"))
	}

	val, err := s.c.ReadDir(root)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.Error(err)
	} else if err != nil {
		return nil, ErrorFileOpen.Error(err)
	} else {
		return val, nil
	}
}

func (s *staticHandler) _fileGet(pathFile string) (fs.FileInfo, io.ReadCloser, liberr.Error) {
	if len(pathFile) < 1 {
		return nil, nil, ErrorParamEmpty.Error(fmt.Errorf("pathfile is empty"))
	}

	if inf, err := s._fileInfo(pathFile); err != nil {
		return nil, nil, err
	} else if inf.Size() >= s._getSize() {
		r, e := s._fileTemp(pathFile)
		return inf, r, e
	} else {
		r, e := s._fileBuff(pathFile)
		return inf, r, e
	}
}

func (s *staticHandler) _fileInfo(pathFile string) (fs.FileInfo, liberr.Error) {
	if pathFile == "" {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("pathfile is empty"))
	}

	var inf fs.FileInfo
	obj, err := s.c.Open(pathFile)

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

func (s *staticHandler) _fileBuff(pathFile string) (io.ReadCloser, liberr.Error) {
	if pathFile == "" {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("pathfile is empty"))
	}

	obj, err := s.c.ReadFile(pathFile)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.Error(err)
	} else if err != nil {
		return nil, ErrorFileOpen.Error(err)
	} else {
		return libbuf.New(bytes.NewBuffer(obj)), nil
	}
}

func (s *staticHandler) _fileTemp(pathFile string) (libfpg.Progress, liberr.Error) {
	if pathFile == "" {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("pathfile is empty"))
	}

	var tmp libfpg.Progress
	obj, err := s.c.Open(pathFile)

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

	return tmp, nil
}

func (s *staticHandler) Has(pathFile string) bool {
	if _, e := s._fileInfo(pathFile); e != nil {
		return false
	} else {
		return true
	}
}

func (s *staticHandler) List(rootPath string) ([]string, liberr.Error) {
	var (
		err error
		res = make([]string, 0)
		lst []string
		ent []fs.DirEntry
		inf fs.FileInfo
	)

	if rootPath == "" {
		for _, p := range s._getBase() {
			inf, err = s._fileInfo(p)
			if err != nil {
				return nil, err.(liberr.Error)
			}

			if !inf.IsDir() {
				res = append(res, p)
				continue
			}

			lst, err = s.List(p)

			if err != nil {
				return nil, err.(liberr.Error)
			}

			res = append(res, lst...)
		}
	} else if ent, err = s._listEmbed(rootPath); err != nil {
		return nil, err.(liberr.Error)
	} else {
		for _, f := range ent {

			if !f.IsDir() {
				res = append(res, path.Join(rootPath, f.Name()))
				continue
			}

			lst, err = s.List(path.Join(rootPath, f.Name()))

			if err != nil {
				return nil, err.(liberr.Error)
			}

			res = append(res, lst...)
		}
	}

	return res, nil
}

func (s *staticHandler) Find(pathFile string) (io.ReadCloser, liberr.Error) {
	_, r, e := s._fileGet(pathFile)
	return r, e
}

func (s *staticHandler) Info(pathFile string) (os.FileInfo, liberr.Error) {
	return s._fileInfo(pathFile)
}

func (s *staticHandler) Temp(pathFile string) (libfpg.Progress, liberr.Error) {
	return s._fileTemp(pathFile)
}

func (s *staticHandler) Map(fct func(pathFile string, inf os.FileInfo) error) liberr.Error {
	var (
		err error
		lst []string
		inf fs.FileInfo
	)

	if lst, err = s.List(""); err != nil {
		return err.(liberr.Error)
	} else {
		for _, f := range lst {
			if inf, err = s._fileInfo(f); err != nil {
				return err.(liberr.Error)
			} else if err = fct(f, inf); err != nil {
				return err.(liberr.Error)
			}
		}
	}

	return nil
}

func (s *staticHandler) UseTempForFileSize(size int64) {
	s._setSize(size)
}
