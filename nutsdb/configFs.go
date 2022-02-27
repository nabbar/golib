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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	libval "github.com/go-playground/validator/v10"
	liberr "github.com/nabbar/golib/errors"
)

const (
	_DefaultFolderData   = "data"
	_DefaultFolderBackup = "backup"
	_DefaultFolderWal    = "wal"
	_DefaultFolderHost   = "host"
)

type NutsDBFolder struct {
	// Working represents the main working folder witch will include sub directories : data, backup, temp...
	// If the base directory is empty, all the sub directory will be absolute directories.
	Base string `mapstructure:"base" json:"base" yaml:"base" toml:"base" validate:"dir,required"`

	// Data represents the sub-dir for the opening database.
	// By default, it will use `data` as sub folder.
	Data string `mapstructure:"sub_data" json:"sub_data" yaml:"sub_data" toml:"sub_data" validate:"printascii,required"`

	// Backup represents the sub-dir with all backup sub-folder.
	// By default, it will use `backup` as sub folder.
	Backup string `mapstructure:"sub_backup" json:"sub_backup" yaml:"sub_backup" toml:"sub_backup" validate:"printascii,required"`

	// Temp represents the sub-dir for temporary file/dir.
	// By default, it will use the system temporary folder.
	Temp string `mapstructure:"sub_temp" json:"sub_temp" yaml:"sub_temp" toml:"sub_temp" validate:"printascii,required"`

	// WalDir represents the sub-dir for cluster negociation.
	// By default, it will use `wal` as sub folder.
	WalDir string `mapstructure:"wal_dir" json:"wal_dir" yaml:"wal_dir" toml:"wal_dir" validate:"printascii,required"`

	// HostDir represents the sub-dir for cluster storage.
	// By default, it will use `host` as sub folder.
	HostDir string `mapstructure:"host_dir" json:"host_dir" yaml:"host_dir" toml:"host_dir" validate:"printascii,required"`

	// LimitNumberBackup represents how many backup will be keep.
	LimitNumberBackup uint8 `mapstructure:"limit_number_backup" json:"limit_number_backup" yaml:"limit_number_backup" toml:"limit_number_backup"`

	// Permission represents the permission apply to folder created.
	// By default, it will use `0755` as permission.
	Permission os.FileMode `mapstructure:"permission" json:"permission" yaml:"permission" toml:"permission"`
}

func (f NutsDBFolder) Validate() liberr.Error {
	err := ErrorValidateConfig.Error(nil)

	if er := libval.New().Struct(f); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.AddParent(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (f NutsDBFolder) getDirectory(base, dir string) (string, liberr.Error) {
	if f.Permission == 0 {
		f.Permission = 0770
	}

	var (
		abs string
		err error
	)

	if len(dir) < 1 {
		return "", nil
	}

	if len(base) > 0 {
		dir = filepath.Join(base, dir)
	}

	if abs, err = filepath.Abs(dir); err != nil {
		return "", ErrorFolderCheck.ErrorParent(err)
	}

	if f.Permission == 0 {
		f.Permission = 0755
	}

	if _, err = os.Stat(abs); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", ErrorFolderCheck.ErrorParent(err)
	} else if err != nil {
		if err = os.MkdirAll(abs, f.Permission); err != nil {
			return "", ErrorFolderCreate.ErrorParent(err)
		}
	}

	return abs, nil
}

func (f NutsDBFolder) GetDirectoryBase() (string, liberr.Error) {
	return f.getDirectory("", f.Base)
}

func (f NutsDBFolder) GetDirectoryData() (string, liberr.Error) {
	if base, err := f.GetDirectoryBase(); err != nil {
		return "", err
	} else if fs, err := f.getDirectory(base, f.Data); err != nil {
		return "", err
	} else if fs == "" {
		return f.getDirectory(base, _DefaultFolderData)
	} else {
		return fs, nil
	}
}

func (f NutsDBFolder) GetDirectoryBackup() (string, liberr.Error) {
	if base, err := f.GetDirectoryBase(); err != nil {
		return "", err
	} else if fs, err := f.getDirectory(base, f.Backup); err != nil {
		return "", err
	} else if fs == "" {
		return f.getDirectory(base, _DefaultFolderBackup)
	} else {
		return fs, nil
	}
}

func (f NutsDBFolder) GetDirectoryWal() (string, liberr.Error) {
	if base, err := f.GetDirectoryBase(); err != nil {
		return "", err
	} else if fs, err := f.getDirectory(base, f.WalDir); err != nil {
		return "", err
	} else if fs == "" {
		return f.getDirectory(base, _DefaultFolderWal)
	} else {
		return fs, nil
	}
}

func (f NutsDBFolder) GetDirectoryHost() (string, liberr.Error) {
	if base, err := f.GetDirectoryBase(); err != nil {
		return "", err
	} else if fs, err := f.getDirectory(base, f.HostDir); err != nil {
		return "", err
	} else if fs == "" {
		return f.getDirectory(base, _DefaultFolderHost)
	} else {
		return fs, nil
	}
}

func (f NutsDBFolder) GetDirectoryTemp() (string, liberr.Error) {
	if base, err := f.GetDirectoryBase(); err != nil {
		return "", err
	} else if fs, err := f.getDirectory(base, f.Temp); err != nil {
		return "", err
	} else if fs == "" {
		return f.getDirectory("", os.TempDir())
	} else {
		return fs, nil
	}
}
