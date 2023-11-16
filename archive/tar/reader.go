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
	"path/filepath"
	"runtime"
	"strings"

	libarc "github.com/nabbar/golib/archive/archive"
	liberr "github.com/nabbar/golib/errors"
	libfpg "github.com/nabbar/golib/file/progress"
)

func GetFile(src, dst libfpg.Progress, filenameContain, filenameRegex string) liberr.Error {

	if _, e := src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
	} else if _, e = dst.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
	}

	r := tar.NewReader(src)

	for {
		h, e := r.Next()
		if e != nil && e == io.EOF {
			return nil
		} else if e != nil {
			return ErrorTarNext.Error(e)
		}

		if h.FileInfo().Mode()&os.ModeType == os.ModeType {
			continue
		}

		f := libarc.NewFileFullPath(h.Name)

		//nolint #nosec
		/* #nosec */
		if f.MatchingFullPath(filenameContain) || f.RegexFullPath(filenameRegex) {
			if _, e = dst.ReadFrom(r); e != nil {
				return ErrorIOCopy.Error(e)
			} else if _, e = dst.Seek(0, io.SeekStart); e != nil {
				return ErrorFileSeek.Error(e)
			} else {
				return nil
			}
		}
	}
}

func GetAll(src io.ReadSeeker, outputFolder string, defaultDirPerm os.FileMode) liberr.Error {

	if _, e := src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
	}

	r := tar.NewReader(src)

	for {
		h, e := r.Next()
		if e != nil && e == io.EOF {
			return nil
		} else if e != nil {
			return ErrorTarNext.Error(e)
		}

		dst := filepath.Join(outputFolder, libarc.CleanPath(strings.Replace(h.Name, ".."+string(filepath.Separator), "", -1)))

		//nolint #nosec
		/* #nosec */
		if err := writeContent(r, h, dst, defaultDirPerm); err != nil {
			return err
		}
	}
}

func writeContent(r io.Reader, h *tar.Header, out string, defaultDirPerm os.FileMode) (err liberr.Error) {
	var (
		e error

		inf = h.FileInfo()
		dst libfpg.Progress
	)

	if e := dirIsExistOrCreate(filepath.Dir(out), defaultDirPerm); e != nil {
		return e
	}

	defer func() {
		if dst != nil {
			if e := dst.Close(); e != nil {
				err = ErrorFileClose.Error(e)
				err.Add(err)
			}
		}
	}()

	if h.Typeflag&tar.TypeDir == tar.TypeDir {
		err = dirIsExistOrCreate(out, h.FileInfo().Mode())
		return
	} else if err = notDirExistCannotClean(out, h.Typeflag, h.Linkname); err != nil {
		return
	} else if h.Typeflag&tar.TypeLink == tar.TypeLink {
		return createLink(out, libarc.CleanPath(h.Linkname), false)
	} else if h.Typeflag&tar.TypeSymlink == tar.TypeSymlink {
		return createLink(out, libarc.CleanPath(h.Linkname), true)
	}

	if dst, e = libfpg.New(out, os.O_RDWR|os.O_CREATE|os.O_TRUNC, inf.Mode()); e != nil {
		return ErrorFileOpen.Error(e)
	} else if _, e = io.Copy(dst, r); e != nil {
		return ErrorIOCopy.Error(e)
	} else if e = dst.Close(); e != nil {
		return ErrorFileClose.Error(e)
	}

	return nil
}

func dirIsExistOrCreate(dirname string, dirPerm os.FileMode) liberr.Error {
	if i, e := os.Stat(dirname); e != nil && os.IsNotExist(e) {
		if e = os.MkdirAll(dirname, dirPerm); e != nil {
			return ErrorDirCreate.Error(e)
		}
	} else if e != nil {
		return ErrorDestinationStat.Error(e)
	} else if !i.IsDir() {
		return ErrorDestinationIsNotDir.Error(nil)
	}

	return nil
}

func notDirExistCannotClean(filename string, flag byte, targetLink string) liberr.Error {
	if strings.EqualFold(runtime.GOOS, "windows") {
		if flag&tar.TypeLink == tar.TypeLink {
			return nil
		} else if flag&tar.TypeSymlink == tar.TypeSymlink {
			return nil
		}
	}

	if _, e := os.Stat(filename); e != nil && os.IsNotExist(e) {
		return nil
	} else if e != nil {
		return ErrorDestinationStat.Error(e)
	} else if flag&tar.TypeLink == tar.TypeLink || flag&tar.TypeSymlink == tar.TypeSymlink {
		if hasFSLink(filename) && compareLinkTarget(filename, targetLink) {
			return nil
		}
	}

	if e := os.Remove(filename); e != nil {
		err := ErrorDestinationRemove.Error(e)
		return err
	}

	return nil
}

func hasFSLink(path string) bool {
	link, _ := filepath.EvalSymlinks(path)

	if link != "" {
		return true
	}

	return false
}

func createLink(link, target string, sym bool) liberr.Error {
	if strings.EqualFold(runtime.GOOS, "windows") {
		return nil
	}

	if _, e := os.Stat(link); e != nil && !os.IsNotExist(e) {
		return ErrorDestinationStat.Error(e)
	} else if e == nil {
		return nil
	} else if compareLinkTarget(link, target) {
		return nil
	}

	if sym {
		err := os.Symlink(libarc.CleanPath(target), libarc.CleanPath(link))
		if err != nil {
			return ErrorLinkCreate.Error(err)
		}
	} else {
		err := os.Link(libarc.CleanPath(target), libarc.CleanPath(link))
		if err != nil {
			return ErrorLinkCreate.Error(err)
		}
	}

	return nil
}

func compareLinkTarget(link, target string) bool {
	var l string

	l, _ = filepath.EvalSymlinks(link)

	if l == "" {
		return false
	}

	return strings.EqualFold(libarc.CleanPath(l), libarc.CleanPath(target))
}
