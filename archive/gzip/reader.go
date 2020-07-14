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

package gzip

import (
	gz "compress/gzip"
	"io"
	"os"

	. "github.com/nabbar/golib/errors"
	//. "github.com/nabbar/golib/logger"

	iou "github.com/nabbar/golib/ioutils"
)

func GetFile(src *os.File, filenameContain, filenameRegex string) (dst *os.File, err Error) {
	if _, e := src.Seek(0, 0); e != nil {
		//ErrorLevel.LogErrorCtx(DebugLevel, "seeking buffer", e)
		return nil, FILE_SEEK.ErrorParent(e)
	}

	r, e := gz.NewReader(src)
	if e != nil {
		//ErrorLevel.LogErrorCtx(DebugLevel, "init gzip reader", e)
		return nil, GZ_READER.ErrorParent(e)
	}

	defer r.Close()

	if t, e := iou.NewTempFile(); e != nil {
		//ErrorLevel.LogErrorCtx(DebugLevel, "init new temporary buffer", e)
		return nil, e
	} else if _, e := io.Copy(t, r); e != nil {
		//ErrorLevel.LogErrorCtx(DebugLevel, "copy buffer from archive reader", e)
		return nil, IO_COPY.ErrorParent(e)
	} else if _, e := t.Seek(0, 0); e != nil {
		//ErrorLevel.LogErrorCtx(DebugLevel, "seeking temp file", e)
		return nil, FILE_SEEK.ErrorParent(e)
	} else {
		return t, nil
	}
}
