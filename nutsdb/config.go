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
	"fmt"

	moncfg "github.com/nabbar/golib/monitor/types"

	libval "github.com/go-playground/validator/v10"
	libclu "github.com/nabbar/golib/cluster"
	liberr "github.com/nabbar/golib/errors"
	"github.com/nutsdb/nutsdb"
)

type Config struct {
	DB        NutsDBOptions `mapstructure:"db" json:"db" yaml:"db" toml:"db" validate:"dive"`
	Cluster   libclu.Config `mapstructure:"cluster" json:"cluster" yaml:"cluster" toml:"cluster" validate:"dive"`
	Directory NutsDBFolder  `mapstructure:"directories" json:"directories" yaml:"directories" toml:"directories" validate:"dive"`
	Monitor   moncfg.Config `mapstructure:"monitor" json:"monitor" yaml:"monitor" toml:"monitor" validate:"dive"`
}

func (c Config) GetConfigFolder() NutsDBFolder {
	return c.Directory
}

func (c Config) GetConfigDB() (nutsdb.Options, liberr.Error) {
	if dir, err := c.Directory.GetDirectoryData(); err != nil {
		return nutsdb.Options{}, err
	} else {
		return c.DB.GetNutsDBOptions(dir), nil
	}
}

func (c Config) GetConfigCluster() (libclu.Config, liberr.Error) {
	cfg := c.Cluster

	if dir, err := c.Directory.GetDirectoryWal(); err != nil {
		return cfg, err
	} else {
		cfg.Node.WALDir = dir
	}

	if dir, err := c.Directory.GetDirectoryHost(); err != nil {
		return cfg, err
	} else {
		cfg.Node.NodeHostDir = dir
	}

	return cfg, nil
}

func (c Config) GetOptions() (Options, liberr.Error) {
	return NewOptions(c.DB, c.Directory)
}

func (c Config) Validate() liberr.Error {
	err := ErrorValidateConfig.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.Add(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (c Config) ValidateDB() liberr.Error {
	return c.DB.Validate()
}

func (c Config) ValidateCluster() liberr.Error {
	return c.Cluster.Validate()
}
