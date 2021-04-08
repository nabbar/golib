/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package nutsdb

import (
	"io"
	"os"
	"path"
	"runtime"

	"github.com/nabbar/golib/archive"
	liberr "github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/ioutils"
	"github.com/xujiajun/nutsdb"
)

type Snapshot interface {
	Prepare(opt Options, db *nutsdb.DB) liberr.Error

	Save(opt Options, writer io.Writer) liberr.Error
	Load(opt Options, reader io.Reader) liberr.Error

	Apply(opt Options) liberr.Error

	Finish()
}

func newSnap() Snapshot {
	s := &snap{}
	runtime.SetFinalizer(s, s.Finish)

	return s
}

func snapCast(i interface{}) (Snapshot, bool) {
	if sp, ok := i.(*snap); ok {
		return sp, true
	} else if ss, ok := i.(snap); ok {
		return &ss, true
	} else if si, ok := i.(Snapshot); ok {
		return si, true
	} else {
		return nil, false
	}
}

type snap struct {
	path string
}

func (s *snap) Prepare(opt Options, db *nutsdb.DB) liberr.Error {
	if dir, err := opt.NewBackupTemp(db); err != nil {
		return err
	} else {
		s.path = dir
	}

	return nil
}

func (s *snap) Save(opt Options, writer io.Writer) liberr.Error {
	var (
		tar string
		err error

		t ioutils.FileProgress
		e liberr.Error
	)

	if tar, e = opt.NewTempFile("tar"); e != nil {
		return e
	}

	defer func() {
		_ = os.Remove(tar)
	}()

	if t, e = ioutils.NewFileProgressPathWrite(tar, false, true, 0664); e != nil {
		return e
	}

	defer func() {
		_ = t.Close()
	}()

	if _, e = archive.CreateArchive(archive.ArchiveTypeTarGzip, t, s.path); e != nil {
		return ErrorFolderArchive.Error(e)
	}

	if _, err = t.Seek(0, io.SeekStart); err != nil {
		return ErrorDatabaseSnapshot.ErrorParent(err)
	}

	if _, err = t.WriteTo(writer); err != nil {
		return ErrorDatabaseSnapshot.ErrorParent(err)
	}

	return nil
}

func (s *snap) Load(opt Options, reader io.Reader) liberr.Error {
	var (
		arc string
		out string
		err error

		a ioutils.FileProgress
		e liberr.Error
	)

	if arc, e = opt.NewTempFile("tar.gz"); e != nil {
		return e
	}

	defer func() {
		_ = os.Remove(arc)
	}()

	if a, e = ioutils.NewFileProgressPathWrite(arc, false, true, 0664); e != nil {
		return e
	}

	defer func() {
		_ = a.Close()
	}()

	if _, err = a.ReadFrom(reader); err != nil {
		return ErrorDatabaseSnapshot.ErrorParent(err)
	}

	if out, e = opt.NewTempFolder(); e != nil {
		return e
	}

	if e = archive.ExtractAll(a, path.Base(arc), out, opt.Permission()); e != nil {
		return ErrorFolderExtract.Error(e)
	}

	s.path = out

	return nil
}

func (s *snap) Apply(opt Options) liberr.Error {
	if e := opt.RestoreBackup(s.path); e != nil {
		return ErrorDatabaseSnapshot.Error(e)
	}

	return nil
}

func (s *snap) Finish() {
	_ = os.RemoveAll(s.path)
}
