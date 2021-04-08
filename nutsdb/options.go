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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/xujiajun/utils/filesystem"

	liberr "github.com/nabbar/golib/errors"

	"github.com/xujiajun/nutsdb"
)

type Options interface {
	NutsDBOptions() nutsdb.Options

	NewBackup(db *nutsdb.DB) (string, liberr.Error)
	NewBackupTemp(db *nutsdb.DB) (string, liberr.Error)

	NewTempFolder() (string, liberr.Error)
	NewTempFile(extension string) (string, liberr.Error)

	CleanBackup() liberr.Error
	Permission() os.FileMode

	RestoreBackup(dir string) liberr.Error
}

func NewOptions(cfgNuts NutsDBOptions, cfgFs NutsDBFolder) (Options, liberr.Error) {
	if _, err := cfgFs.GetDirectoryBase(); err != nil {
		return nil, err
	}

	o := &options{
		limit: cfgFs.LimitNumberBackup,
		perm:  cfgFs.Permission,
	}

	if fs, err := cfgFs.GetDirectoryData(); err != nil {
		return nil, err
	} else {
		o.dirs.data = fs
	}

	if fs, err := cfgFs.GetDirectoryBackup(); err != nil {
		return nil, err
	} else {
		o.dirs.backup = fs
	}

	if fs, err := cfgFs.GetDirectoryTemp(); err != nil {
		return nil, err
	} else {
		o.dirs.temp = fs
	}

	o.nuts = cfgNuts.GetNutsDBOptions(o.dirs.data)

	return o, nil
}

type options struct {
	nuts nutsdb.Options
	dirs struct {
		data   string
		backup string
		temp   string
	}
	limit uint8
	perm  os.FileMode
}

func (o options) NutsDBOptions() nutsdb.Options {
	return o.nuts
}

func (o options) NewBackup(db *nutsdb.DB) (string, liberr.Error) {
	fld := o.getBackupDirName()

	if e := os.MkdirAll(filepath.Join(o.dirs.backup, fld), o.perm); e != nil {
		return "", ErrorFolderCreate.ErrorParent(e)
	} else if err := o.newBackupDir(fld, db); err != nil {
		return "", err
	} else {
		return fld, nil
	}
}

func (o options) NewBackupTemp(db *nutsdb.DB) (string, liberr.Error) {
	if fld, err := o.NewTempFolder(); err != nil {
		return "", err
	} else if err = o.newBackupDir(fld, db); err != nil {
		return "", err
	} else {
		return fld, nil
	}
}

func (o options) NewTempFolder() (string, liberr.Error) {
	if p, e := ioutil.TempDir(o.dirs.temp, o.getTempPrefix()); e != nil {
		return "", ErrorFolderCreate.ErrorParent(e)
	} else {
		_ = os.Chmod(p, o.perm)
		return p, nil
	}
}

func (o options) NewTempFile(extension string) (string, liberr.Error) {
	pattern := o.getTempPrefix() + "-*"

	if extension != "" {
		pattern = pattern + "." + extension
	}

	if file, e := ioutil.TempFile(o.dirs.temp, pattern); e != nil {
		return "", ErrorFolderCreate.ErrorParent(e)
	} else {
		p := file.Name()
		_ = file.Close()

		return p, nil
	}
}

func (o options) CleanBackup() liberr.Error {
	panic("implement me")
}

func (o options) Permission() os.FileMode {
	return o.perm
}

func (o options) RestoreBackup(dir string) liberr.Error {
	if err := os.RemoveAll(o.dirs.data); err != nil {
		return ErrorFolderDelete.ErrorParent(err)
	} else if err = filesystem.CopyDir(dir, o.dirs.data); err != nil {
		return ErrorFolderCopy.ErrorParent(err)
	} else {
		_ = os.Chmod(o.dirs.data, o.perm)
	}

	return nil
}

/// private

func (o options) newBackupDir(dir string, db *nutsdb.DB) liberr.Error {
	if err := db.Backup(dir); err != nil {
		return ErrorDatabaseBackup.ErrorParent(err)
	}

	return nil
}

func (o options) getTempPrefix() string {
	b := make([]byte, 64)

	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]

	//nolint #nosec
	/* #nosec */
	n, _ := strconv.ParseUint(string(b), 10, 64)

	return fmt.Sprintf("%d", n)
}

func (o options) getBackupDirName() string {
	part := strings.Split(time.Now().Format(time.RFC3339), "T")
	dt := part[0]

	part = strings.Split(part[1], "Z")
	tm := strings.Replace(part[0], ":", "-", -1)
	tz := strings.Replace(part[1], ":", "", -1)

	return dt + "T" + tm + "Z" + tz
}
