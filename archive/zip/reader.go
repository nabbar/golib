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

package zip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	arcmod "github.com/nabbar/golib/archive/archive"
	liberr "github.com/nabbar/golib/errors"
	libiot "github.com/nabbar/golib/ioutils"
)

func GetFile(src, dst libiot.FileProgress, filenameContain, filenameRegex string) liberr.Error {
	var (
		arc *zip.Reader
		inf os.FileInfo
		err error
	)

	if _, err = src.Seek(0, io.SeekStart); err != nil {
		return ErrorFileSeek.ErrorParent(err)
	} else if _, err = dst.Seek(0, io.SeekStart); err != nil {
		return ErrorFileSeek.ErrorParent(err)
	} else if inf, err = src.FileStat(); err != nil {
		return ErrorFileStat.ErrorParent(err)
	} else if arc, err = zip.NewReader(src, inf.Size()); err != nil {
		return ErrorZipOpen.ErrorParent(err)
	}

	for _, f := range arc.File {
		if f.Mode()&os.ModeType == os.ModeType {
			continue
		}

		z := arcmod.NewFileFullPath(f.Name)

		if z.MatchingFullPath(filenameContain) || z.RegexFullPath(filenameRegex) {
			if f == nil {
				continue
			}

			var (
				r io.ReadCloser
				e error
			)

			if r, e = f.Open(); e != nil {
				//logger.ErrorLevel.LogErrorCtx(logger.DebugLevel, "open zipped file reader", err)
				return ErrorZipFileOpen.ErrorParent(e)
			}

			defer func() {
				_ = r.Close()
			}()

			//nolint #nosec
			/* #nosec */
			if _, e = dst.ReadFrom(r); e != nil {
				//logger.ErrorLevel.LogErrorCtx(logger.DebugLevel, "copy buffer from archive reader", err)
				return ErrorIOCopy.ErrorParent(e)
			}

			if _, e = dst.Seek(0, io.SeekStart); e != nil {
				//logger.ErrorLevel.LogErrorCtx(logger.DebugLevel, "seeking temp file", err)
				return ErrorFileSeek.ErrorParent(e)
			}

			return nil
		}
	}

	return nil
}

func GetAll(src libiot.FileProgress, outputFolder string, defaultDirPerm os.FileMode) liberr.Error {
	var (
		r *zip.Reader
		i os.FileInfo
		e error
	)

	if _, e = src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.ErrorParent(e)
	} else if i, e = src.FileStat(); e != nil {
		return ErrorFileStat.ErrorParent(e)
	} else if r, e = zip.NewReader(src, i.Size()); e != nil {
		return ErrorZipOpen.ErrorParent(e)
	}

	for _, f := range r.File {
		if f == nil {
			continue
		}

		//nolint #nosec
		/* #nosec */
		if err := writeContent(f, filepath.Join(outputFolder, arcmod.CleanPath(f.Name)), defaultDirPerm); err != nil {
			return err
		}
	}

	return nil
}

func writeContent(f *zip.File, out string, defaultDirPerm os.FileMode) (err liberr.Error) {
	var (
		dst libiot.FileProgress
		inf = f.FileInfo()

		r io.ReadCloser
		e error
	)

	if err = dirIsExistOrCreate(filepath.Dir(out), defaultDirPerm); err != nil {
		return
	}

	defer func() {
		if dst != nil {
			if e = dst.Close(); e != nil {
				err = ErrorFileClose.ErrorParent(e)
				err.AddParentError(err)
			}
		}
		if r != nil {
			if e = r.Close(); e != nil {
				err = ErrorZipFileClose.Error(err)
			}
		}
	}()

	if inf.IsDir() {
		err = dirIsExistOrCreate(out, inf.Mode())
		return
	} else if inf.Mode()&os.ModeSymlink == os.ModeSymlink {
		return nil
	} else if err = notDirExistCannotClean(out); err != nil {
		return
	}

	if dst, err = libiot.NewFileProgressPathWrite(out, true, true, inf.Mode()); err != nil {
		return ErrorFileOpen.Error(err)
	} else if r, e = f.Open(); e != nil {
		return ErrorZipFileOpen.ErrorParent(e)
	}

	//nolint #nosec
	/* #nosec */
	if _, e = io.Copy(dst, r); e != nil {
		return ErrorIOCopy.ErrorParent(e)
	} else if e = dst.Close(); e != nil {
		return ErrorFileClose.ErrorParent(e)
	}

	return nil
}

func dirIsExistOrCreate(dirname string, dirPerm os.FileMode) liberr.Error {
	if i, e := os.Stat(filepath.Dir(dirname)); e != nil && os.IsNotExist(e) {
		if e = os.MkdirAll(filepath.Dir(dirname), dirPerm); e != nil {
			return ErrorDirCreate.ErrorParent(e)
		}
	} else if e != nil {
		return ErrorDestinationStat.ErrorParent(e)
	} else if !i.IsDir() {
		return ErrorDestinationIsNotDir.Error(nil)
	}

	return nil
}

func notDirExistCannotClean(filename string) liberr.Error {
	if i, e := os.Stat(filename); e != nil && !os.IsNotExist(e) {
		return ErrorDestinationStat.ErrorParent(e)
	} else if e == nil && i.IsDir() {
		return ErrorDestinationIsDir.Error(nil)
	} else if e == nil {
		if e = os.Remove(filename); e != nil {
			return ErrorDestinationRemove.ErrorParent(e)
		}
	}
	return nil
}
