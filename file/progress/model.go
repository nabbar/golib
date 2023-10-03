/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package progress

import (
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
)

type progress struct {
	fos *os.File

	b *atomic.Int32

	fi *atomic.Value
	fe *atomic.Value
	fr *atomic.Value
}

func (o *progress) SetBufferSize(size int32) {
	o.b.Store(size)
}

func (o *progress) getBufferSize(size int) int {
	if size > 0 {
		return size
	} else if o == nil {
		return DefaultBuffSize
	}

	i := o.b.Load()
	if i < 1024 {
		return DefaultBuffSize
	} else {
		return int(i)
	}
}

func (o *progress) Path() string {
	return filepath.Clean(o.fos.Name())
}

func (o *progress) Stat() (os.FileInfo, error) {
	if o == nil || o.fos == nil {
		return nil, ErrorNilPointer.Error(nil)
	}

	if i, e := o.fos.Stat(); e != nil {
		return i, ErrorIOFileStat.Error(e)
	} else {
		return i, nil
	}
}

func (o *progress) SizeBOF() (size int64, err error) {
	if o == nil || o.fos == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	return o.seek(0, io.SeekCurrent)
}

func (o *progress) SizeEOF() (size int64, err error) {
	if o == nil || o.fos == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	var (
		e error
		a int64 // origin
		b int64 // eof
	)

	if a, e = o.seek(0, io.SeekCurrent); e != nil {
		return 0, e
	} else if b, e = o.seek(0, io.SeekEnd); e != nil {
		return 0, e
	} else if _, e = o.seek(a, io.SeekStart); e != nil {
		return 0, e
	} else {
		return b - a, nil
	}
}

func (o *progress) Truncate(size int64) error {
	if o == nil || o.fos == nil {
		return ErrorNilPointer.Error(nil)
	}

	e := o.fos.Truncate(size)
	o.reset()

	return e
}

func (o *progress) Sync() error {
	if o == nil || o.fos == nil {
		return ErrorNilPointer.Error(nil)
	}

	return o.fos.Sync()
}
