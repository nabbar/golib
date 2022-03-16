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

package ioutils

import (
	"bufio"
	"context"
	"io"
	"os"
	"path"

	"github.com/nabbar/golib/errors"
)

const buffSize = 32 * 1024 // see io.copyBuffer

type FileProgress interface {
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

	GetByteReader() io.ByteReader
	GetByteWriter() io.ByteWriter

	SetContext(ctx context.Context)
	SetIncrement(increment func(size int64))
	SetFinish(finish func())
	SetReset(reset func(size, current int64))
	ResetMax(max int64)

	FilePath() string
	FileStat() (os.FileInfo, errors.Error)

	SizeToEOF() (size int64, err errors.Error)

	NewFilePathMode(filepath string, mode int, perm os.FileMode) (FileProgress, errors.Error)
	NewFilePathWrite(filepath string, create, overwrite bool, perm os.FileMode) (FileProgress, errors.Error)
	NewFilePathRead(filepath string, perm os.FileMode) (FileProgress, errors.Error)
	NewFileTemp() (FileProgress, errors.Error)

	CloseNoClean() error
}

func NewFileProgress(file *os.File) FileProgress {
	return &fileProgress{
		fs: file,
		fc: nil,
		ff: nil,
		fr: nil,
		t:  false,
		x:  nil,
	}
}

func NewFileProgressTemp() (FileProgress, errors.Error) {
	if fs, e := NewTempFile(); e != nil {
		return nil, e
	} else {
		return &fileProgress{
			fs: fs,
			fc: nil,
			ff: nil,
			fr: nil,
			t:  true,
			x:  nil,
		}, nil
	}
}

func NewFileProgressPathOpen(filepath string) (FileProgress, errors.Error) {
	//nolint #gosec
	/* #nosec */
	if f, err := os.Open(filepath); err != nil {
		return nil, ErrorIOFileOpen.ErrorParent(err)
	} else {
		return NewFileProgress(f), nil
	}
}

func NewFileProgressPathMode(filepath string, mode int, perm os.FileMode) (FileProgress, errors.Error) {
	//nolint #nosec
	/* #nosec */
	if f, err := os.OpenFile(filepath, mode, perm); err != nil {
		return nil, ErrorIOFileOpen.ErrorParent(err)
	} else {
		return NewFileProgress(f), nil
	}
}

func NewFileProgressPathWrite(filepath string, create, overwrite bool, perm os.FileMode) (FileProgress, errors.Error) {
	mode := os.O_RDWR | os.O_TRUNC

	if _, err := os.Stat(filepath); err != nil && os.IsNotExist(err) && create {
		mode = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	} else if err != nil {
		return nil, ErrorIOFileStat.ErrorParent(err)
	}

	return NewFileProgressPathMode(filepath, mode, perm)
}

func NewFileProgressPathRead(filepath string, perm os.FileMode) (FileProgress, errors.Error) {
	mode := os.O_RDONLY

	if _, err := os.Stat(filepath); err != nil && !os.IsExist(err) {
		return nil, ErrorIOFileStat.ErrorParent(err)
	}

	return NewFileProgressPathMode(filepath, mode, perm)
}

type fileProgress struct {
	t  bool
	x  context.Context
	fs *os.File
	fc func(size int64)
	ff func()
	fr func(size, current int64)
}

func (f *fileProgress) NewFilePathMode(filepath string, mode int, perm os.FileMode) (FileProgress, errors.Error) {
	if f == nil {
		return nil, ErrorNilPointer.Error(nil)
	}

	if fs, e := NewFileProgressPathMode(filepath, mode, perm); e != nil {
		return nil, e
	} else {
		fs.SetContext(f.x)
		fs.SetFinish(f.ff)
		fs.SetIncrement(f.fc)
		fs.SetReset(f.fr)
		return fs, nil
	}
}

func (f *fileProgress) NewFilePathWrite(filepath string, create, overwrite bool, perm os.FileMode) (FileProgress, errors.Error) {
	if f == nil {
		return nil, ErrorNilPointer.Error(nil)
	}

	if fs, e := NewFileProgressPathWrite(filepath, create, overwrite, perm); e != nil {
		return nil, e
	} else {
		fs.SetContext(f.x)
		fs.SetFinish(f.ff)
		fs.SetIncrement(f.fc)
		fs.SetReset(f.fr)
		return fs, nil
	}
}

func (f *fileProgress) NewFilePathRead(filepath string, perm os.FileMode) (FileProgress, errors.Error) {
	if f == nil {
		return nil, ErrorNilPointer.Error(nil)
	}

	if fs, e := NewFileProgressPathRead(filepath, perm); e != nil {
		return nil, e
	} else {
		fs.SetContext(f.x)
		fs.SetFinish(f.ff)
		fs.SetIncrement(f.fc)
		fs.SetReset(f.fr)
		return fs, nil
	}
}

func (f *fileProgress) NewFileTemp() (FileProgress, errors.Error) {
	if f == nil {
		return nil, ErrorNilPointer.Error(nil)
	}

	if fs, e := NewFileProgressTemp(); e != nil {
		return nil, e
	} else {
		fs.SetContext(f.x)
		fs.SetFinish(f.ff)
		fs.SetIncrement(f.fc)
		fs.SetReset(f.fr)
		return fs, nil
	}
}

func (f *fileProgress) SetContext(ctx context.Context) {
	if f != nil {
		f.x = ctx
	}
}

func (f *fileProgress) checkContext() error {
	if f == nil {
		return ErrorNilPointer.Error(nil)
	}

	if f.x != nil && f.x.Err() != nil {
		return f.x.Err()
	}

	return nil
}

func (f *fileProgress) FilePath() string {
	if f == nil || f.fs == nil {
		return ""
	}

	return path.Clean(f.fs.Name())
}

func (f *fileProgress) FileStat() (os.FileInfo, errors.Error) {
	if f == nil {
		return nil, ErrorNilPointer.Error(nil)
	}

	if i, e := f.fs.Stat(); e != nil {
		return i, ErrorIOFileStat.ErrorParent(e)
	} else {
		return i, nil
	}
}

func (f *fileProgress) SizeToEOF() (size int64, err errors.Error) {
	if f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	if a, e := f.Seek(0, io.SeekCurrent); e != nil {
		return 0, ErrorIOFileSeek.ErrorParent(e)
	} else if b, e := f.Seek(0, io.SeekEnd); e != nil {
		return 0, ErrorIOFileSeek.ErrorParent(e)
	} else if _, e := f.Seek(a, io.SeekStart); e != nil {
		return 0, ErrorIOFileSeek.ErrorParent(e)
	} else {
		return b - a, nil
	}
}

func (f *fileProgress) SetIncrement(increment func(size int64)) {
	if f != nil {
		f.fc = increment
	}
}

func (f *fileProgress) SetFinish(finish func()) {
	if f != nil {
		f.ff = finish
	}
}

func (f *fileProgress) SetReset(reset func(size, current int64)) {
	if f != nil {
		f.fr = reset
	}
}

func (f *fileProgress) GetByteReader() io.ByteReader {
	if f == nil {
		return nil
	}

	return bufio.NewReaderSize(f, buffSize)
}

func (f *fileProgress) GetByteWriter() io.ByteWriter {
	if f == nil {
		return nil
	}

	return bufio.NewWriterSize(f, buffSize)
}

func (f *fileProgress) ReadFrom(r io.Reader) (n int64, err error) {
	if f == nil || r == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	size := buffSize
	if l, ok := r.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	buf := make([]byte, size)

	for {
		if err = f.checkContext(); err != nil {
			break
		}

		// copy from bufio CopyBuff
		nr, er := r.Read(buf)

		if nr > 0 {
			nw, ew := f.Write(buf[0:nr])
			if nw > 0 {
				n += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	f.finish()
	return n, err
}

func (f *fileProgress) WriteTo(w io.Writer) (n int64, err error) {
	if f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	buf := make([]byte, buffSize)

	for {
		if err = f.checkContext(); err != nil {
			break
		}

		// copy from bufio CopyBuff
		nr, er := f.Read(buf)

		if nr > 0 {
			nw, ew := w.Write(buf[0:nr])
			if nw > 0 {
				n += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	f.finish()
	return n, err
}

//nolint #unparam
func (f *fileProgress) increment(n int64, size int) {
	n += int64(size)

	if f != nil && f.fc != nil && n > 0 {
		f.fc(n)
	}
}

func (f *fileProgress) finish() {
	if f != nil && f.ff != nil {
		f.ff()
	}
}

func (f *fileProgress) reset(pos int64) {
	if f.fr != nil {
		if i, e := f.FileStat(); e != nil {
			return
		} else if i.Size() != 0 {
			f.fr(i.Size(), pos)
		}
	}
}

func (f *fileProgress) ResetMax(max int64) {
	if f.fr != nil {
		f.fr(max, 0)
	}
}

func (f *fileProgress) Read(p []byte) (n int, err error) {
	if f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	n, err = f.fs.Read(p)
	//logger.DebugLevel.Logf("Loading from '%s' '%d' bytes, with err '%v'", f.FilePath(), n, err)

	if err != nil && err == io.EOF {
		f.finish()
	} else if err == nil {
		f.increment(0, n)
	}

	return
}

func (f *fileProgress) ReadAt(p []byte, off int64) (n int, err error) {
	if f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	n, err = f.fs.ReadAt(p, off)

	if err != nil && err == io.EOF {
		f.finish()
	} else if err == nil {
		f.increment(0, n)
	}

	return
}

func (f *fileProgress) Write(p []byte) (n int, err error) {
	if f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	n, err = f.fs.Write(p)
	//logger.DebugLevel.Logf("Writing from '%s' '%d' bytes, with err '%v'", f.FilePath(), n, err)

	if err != nil && err == io.EOF {
		f.finish()
	} else if err == nil {
		f.increment(0, n)
	}

	return
}

func (f *fileProgress) WriteAt(p []byte, off int64) (n int, err error) {
	if f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	n, err = f.fs.WriteAt(p, off)

	if err != nil && err == io.EOF {
		f.finish()
	} else if err == nil {
		f.increment(0, n)
	}

	return
}

func (f *fileProgress) Seek(offset int64, whence int) (n int64, err error) {
	if f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	n, err = f.fs.Seek(offset, whence)

	if err != nil && err != io.EOF {
		return
	}

	f.reset(n)

	return
}

func (f *fileProgress) WriteString(s string) (n int, err error) {
	if f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	n, err = f.fs.WriteString(s)

	if err != nil && err == io.EOF {
		f.finish()
	} else if err == nil {
		f.increment(0, n)
	}

	return
}

func (f *fileProgress) clean(err error) error {
	f.ff = nil
	f.fr = nil
	f.fc = nil
	f.fs = nil
	return err
}

func (f *fileProgress) Close() error {
	if f == nil || f.fs == nil {
		return nil
	} else if f.t {
		return f.clean(DelTempFile(f.fs))
	} else if f.fs != nil {
		return f.clean(f.fs.Close())
	}

	return nil
}

func (f *fileProgress) CloseNoClean() error {
	return f.clean(f.fs.Close())
}
