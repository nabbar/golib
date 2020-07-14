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

	. "github.com/nabbar/golib/errors"
	//. "github.com/nabbar/golib/logger"

	iou "github.com/nabbar/golib/ioutils"

	"github.com/nabbar/golib/archive/archive"
)

func GetFile(src *os.File, filenameContain, filenameRegex string) (dst *os.File, err Error) {
	location := iou.GetTempFilePath(src)

	if e := src.Close(); err != nil {
		//ErrorLevel.LogErrorCtx(DebugLevel, "trying to close temp file", err)
		return dst, FILE_CLOSE.ErrorParent(e)
	}

	return getFile(location, filenameContain, filenameRegex)
}

func getFile(src string, filenameContain, filenameRegex string) (dst *os.File, err Error) {
	var (
		r *zip.ReadCloser
		e error
	)

	if r, e = zip.OpenReader(src); e != nil {
		//ErrorLevel.LogErrorCtx(DebugLevel, "trying to open zip file", e)
		return nil, ZIP_OPEN.ErrorParent(e)
	}

	defer func() {
		_ = r.Close()
	}()

	for _, f := range r.File {
		if f.Mode()&os.ModeType == os.ModeType {
			continue
		}

		z := archive.NewFileFullPath(f.Name)
		if z.MatchingFullPath(filenameContain) || z.RegexFullPath(filenameRegex) {
			return extratFile(f)
		}
	}

	return nil, nil
}

func extratFile(f *zip.File) (dst *os.File, err Error) {
	var (
		r io.ReadCloser
		e error
	)

	if r, e = f.Open(); e != nil {
		//ErrorLevel.LogErrorCtx(DebugLevel, "open zipped file reader", err)
		return dst, FILE_OPEN.ErrorParent(e)
	}

	defer func() {
		_ = r.Close()
	}()

	if dst, err = iou.NewTempFile(); err != nil {
		//ErrorLevel.LogErrorCtx(DebugLevel, "init new temporary buffer", err)
		return
	} else if _, e = io.Copy(dst, r); e != nil {
		//ErrorLevel.LogErrorCtx(DebugLevel, "copy buffer from archive reader", err)
		return dst, IO_COPY.ErrorParent(e)
	}

	_, e = dst.Seek(0, 0)
	//ErrorLevel.LogErrorCtx(DebugLevel, "seeking temp file", err)
	return dst, FILE_SEEK.ErrorParent(e)
}
