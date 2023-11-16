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
	"sync/atomic"
)

const DefaultBuffSize = 32 * 1024 // see io.copyBuffer

type FctIncrement func(size int64)
type FctReset func(size, current int64)
type FctEOF func()

type GenericIO interface {
	io.ReadCloser
	io.ReadSeeker
	io.ReadWriteCloser
	io.ReadWriteSeeker
	io.WriteCloser
	io.WriteSeeker
	io.Reader
	io.ReaderFrom
	io.ReaderAt
	io.Writer
	io.WriterAt
	io.WriterTo
	io.Seeker
	io.StringWriter
	io.Closer
	io.ByteReader
	io.ByteWriter
}

type File interface {
	CloseDelete() error

	Path() string
	Stat() (os.FileInfo, error)

	SizeBOF() (size int64, err error)
	SizeEOF() (size int64, err error)

	Truncate(size int64) error
	Sync() error
}

type Progress interface {
	GenericIO
	File

	RegisterFctIncrement(fct FctIncrement)
	RegisterFctReset(fct FctReset)
	RegisterFctEOF(fct FctEOF)
	SetBufferSize(size int32)
	SetRegisterProgress(f Progress)

	Reset(max int64)
}

func New(name string, flags int, perm os.FileMode) (Progress, error) {
	// #nosec
	f, e := os.OpenFile(name, flags, perm)

	if e != nil {
		return nil, e
	} else {
		return &progress{
			fos: f,
			b:   new(atomic.Int32),
			fi:  new(atomic.Value),
			fe:  new(atomic.Value),
			fr:  new(atomic.Value),
		}, nil
	}
}

func Unique(basePath, pattern string) (Progress, error) {
	// #nosec
	f, e := os.CreateTemp(basePath, pattern)

	if e != nil {
		return nil, e
	} else {
		return &progress{
			fos: f,
			b:   new(atomic.Int32),
			fi:  new(atomic.Value),
			fe:  new(atomic.Value),
			fr:  new(atomic.Value),
		}, nil
	}
}

func Temp(pattern string) (Progress, error) {
	// #nosec
	f, e := os.CreateTemp("", pattern)

	if e != nil {
		return nil, e
	} else {
		return &progress{
			fos: f,
			b:   new(atomic.Int32),
			fi:  new(atomic.Value),
			fe:  new(atomic.Value),
			fr:  new(atomic.Value),
		}, nil
	}
}

func Open(name string) (Progress, error) {
	// #nosec
	f, e := os.Open(name)

	if e != nil {
		return nil, e
	} else {
		return &progress{
			fos: f,
			b:   new(atomic.Int32),
			fi:  new(atomic.Value),
			fe:  new(atomic.Value),
			fr:  new(atomic.Value),
		}, nil
	}
}

func Create(name string) (Progress, error) {
	// #nosec
	f, e := os.Create(name)

	if e != nil {
		return nil, e
	} else {
		return &progress{
			fos: f,
			b:   new(atomic.Int32),
			fi:  new(atomic.Value),
			fe:  new(atomic.Value),
			fr:  new(atomic.Value),
		}, nil
	}
}
