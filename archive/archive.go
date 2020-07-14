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
	"os"

	"github.com/nabbar/golib/archive/bz2"
	"github.com/nabbar/golib/archive/gzip"
	"github.com/nabbar/golib/archive/tar"
	"github.com/nabbar/golib/archive/zip"

	. "github.com/nabbar/golib/errors"
	//. "github.com/nabbar/golib/logger"

	iou "github.com/nabbar/golib/ioutils"
)

type ArchiveType uint8

func ExtractFile(src *os.File, fileNameContain, fileNameRegex string) (*os.File, Error) {
	loc := src.Name()

	if dst, err := bz2.GetFile(src, fileNameContain, fileNameRegex); err == nil {
		//DebugLevel.Log("try deleting source archive...")
		if err = iou.DelTempFile(src); err != nil {
			return nil, err
		}
		//DebugLevel.Log("try another archive...")
		return ExtractFile(dst, fileNameContain, fileNameRegex)
	}

	if dst, err := gzip.GetFile(src, fileNameContain, fileNameRegex); err == nil {
		//DebugLevel.Log("try deleting source archive...")
		if err = iou.DelTempFile(src); err != nil {
			return nil, err
		}
		//DebugLevel.Log("try another archive...")
		return ExtractFile(dst, fileNameContain, fileNameRegex)
	}

	if dst, err := tar.GetFile(src, fileNameContain, fileNameRegex); err == nil {
		//DebugLevel.Log("try deleting source archive...")
		if err = iou.DelTempFile(src); err != nil {
			return nil, err
		}
		//DebugLevel.Log("try another archive...")
		return ExtractFile(dst, fileNameContain, fileNameRegex)
	}

	if dst, err := zip.GetFile(src, fileNameContain, fileNameRegex); err == nil {
		//DebugLevel.Log("try deleting source archive...")
		if err = iou.DelTempFile(src); err != nil {
			return nil, err
		}
		//DebugLevel.Log("try another archive...")
		return ExtractFile(dst, fileNameContain, fileNameRegex)
	}

	var (
		err error
	)

	if _, err = src.Seek(0, 0); err != nil {
		e1 := FILE_SEEK.ErrorParent(err)
		// #nosec
		if src, err = os.Open(loc); err != nil {
			//ErrorLevel.LogErrorCtx(DebugLevel, "reopening file", err)
			e2 := FILE_OPEN.ErrorParent(err)
			e2.AddParentError(e1)
			return nil, e2
		}
	}

	return src, nil
}
