//go:build !386 && !arm && !mips && !mipsle
// +build !386,!arm,!mips,!mipsle

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
	"strings"

	arcarc "github.com/nabbar/golib/archive/archive"
	arctps "github.com/nabbar/golib/archive/archive/types"

	libarc "github.com/nabbar/golib/archive"
	arccmp "github.com/nabbar/golib/archive/compress"
	liberr "github.com/nabbar/golib/errors"
	"github.com/nutsdb/nutsdb"
)

type Snapshot interface {
	Prepare(opt Options, db *nutsdb.DB) liberr.Error

	Save(opt Options, writer io.Writer) liberr.Error
	Load(opt Options, reader io.Reader) liberr.Error

	Apply(opt Options) liberr.Error

	Finish()
}

func newSnap() Snapshot {
	return &snap{}
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
		e error
		g io.WriteCloser
		t arctps.Writer
		f = func(str string) string {
			return strings.TrimPrefix(str, s.path)
		}
	)

	defer func() {
		if t != nil {
			_ = t.Close()
		}
		if g != nil {
			_ = g.Close()
		}
	}()

	if g, e = arccmp.Gzip.Writer(libarc.NopWriteCloser(writer)); e != nil {
		return ErrorDatabaseSnapshot.Error(e)
	} else if t, e = arcarc.Tar.Writer(g); e != nil {
		return ErrorDatabaseSnapshot.Error(e)
	} else if e = t.FromPath(s.path, "", f); e != nil {
		return ErrorDatabaseSnapshot.Error(e)
	}

	return nil
}

func (s *snap) Load(opt Options, reader io.Reader) liberr.Error {
	var (
		e error
		d string
	)

	defer func() {
	}()

	if d, e = opt.NewTempFolder(); e != nil {
		return ErrorDatabaseSnapshot.Error(e)
	} else if e = libarc.ExtractAll(io.NopCloser(reader), "unknown", d); e != nil {
		return ErrorDatabaseSnapshot.Error(e)
	} else {
		s.path = d
	}

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
