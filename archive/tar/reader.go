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

package tar

import (
	"archive/tar"
	"io"
	"os"
	"path"
	"syscall"

	"github.com/nabbar/golib/archive/archive"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/ioutils"
)

func GetFile(src, dst ioutils.FileProgress, filenameContain, filenameRegex string) errors.Error {

	if _, e := src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.ErrorParent(e)
	} else if _, e = dst.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.ErrorParent(e)
	}

	r := tar.NewReader(src)

	for {
		h, e := r.Next()
		if e != nil && e == io.EOF {
			return nil
		} else if e != nil {
			return ErrorTarNext.ErrorParent(e)
		}

		if h.FileInfo().Mode()&os.ModeType == os.ModeType {
			continue
		}

		f := archive.NewFileFullPath(h.Name)

		//nolint #nosec
		/* #nosec */
		if f.MatchingFullPath(filenameContain) || f.RegexFullPath(filenameRegex) {
			if _, e := dst.ReadFrom(r); e != nil {
				return ErrorIOCopy.ErrorParent(e)
			} else if _, e := dst.Seek(0, io.SeekStart); e != nil {
				return ErrorFileSeek.ErrorParent(e)
			} else {
				return nil
			}
		}
	}
}

func GetAll(src io.ReadSeeker, outputFolder string, defaultDirPerm os.FileMode) errors.Error {

	if _, e := src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.ErrorParent(e)
	}

	r := tar.NewReader(src)

	for {
		h, e := r.Next()
		if e != nil && e == io.EOF {
			return nil
		} else if e != nil {
			return ErrorTarNext.ErrorParent(e)
		}

		//nolint #nosec
		/* #nosec */
		if err := writeContent(r, h, path.Join(outputFolder, h.Name), defaultDirPerm); err != nil {
			return err
		}
	}
}

func writeContent(r io.Reader, h *tar.Header, out string, defaultDirPerm os.FileMode) (err errors.Error) {
	var (
		inf = h.FileInfo()
		dst ioutils.FileProgress
	)

	if e := dirIsExistOrCreate(path.Dir(out), defaultDirPerm); e != nil {
		return e
	}

	defer func() {
		if dst != nil {
			if e := dst.Close(); e != nil {
				err = ErrorFileClose.ErrorParent(e)
				err.AddParentError(err)
			}
		}
	}()

	if h.Typeflag&tar.TypeDir == tar.TypeDir {
		err = dirIsExistOrCreate(out, h.FileInfo().Mode())
		return
	} else if err = notDirExistCannotClean(out); err != nil {
		return
	} else if h.Typeflag&tar.TypeLink == tar.TypeLink {
		e := os.Link(h.Linkname, out)
		if e != nil {
			err = ErrorLinkCreate.ErrorParent(e)
		}
		return
	} else if h.Typeflag&tar.TypeSymlink == tar.TypeSymlink {
		e := os.Symlink(h.Linkname, out)
		if e != nil {
			err = ErrorSymLinkCreate.ErrorParent(e)
		}
		return
	}

	if dst, err = ioutils.NewFileProgressPathWrite(out, true, true, inf.Mode()); err != nil {
		return ErrorFileOpen.Error(err)
	} else if _, e := io.Copy(dst, r); e != nil {
		return ErrorIOCopy.ErrorParent(e)
	} else if e := dst.Close(); e != nil {
		return ErrorFileClose.ErrorParent(e)
	}

	return nil
}

func dirIsExistOrCreate(dirname string, dirPerm os.FileMode) errors.Error {
	if i, e := os.Stat(dirname); e != nil && os.IsNotExist(e) {
		if e = os.MkdirAll(dirname, dirPerm); e != nil {
			return ErrorDirCreate.ErrorParent(e)
		}
	} else if e != nil {
		return ErrorDestinationStat.ErrorParent(e)
	} else if !i.IsDir() {
		return ErrorDestinationIsNotDir.Error(nil)
	}

	return nil
}

func notDirExistCannotClean(filename string) errors.Error {
	if _, e := os.Stat(filename); e != nil && os.IsNotExist(e) {
		return nil
	} else if e != nil {
		return ErrorDestinationStat.ErrorParent(e)
	} else if e = os.Remove(filename); e != nil {
		err := ErrorDestinationRemove.ErrorParent(e)

		if e = syscall.Rmdir(filename); e != nil {
			err.AddParentError(ErrorDestinationIsDir.ErrorParent(e))
		}

		return err
	}

	return nil
}
