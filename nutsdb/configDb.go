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

	libval "github.com/go-playground/validator/v10"
	liberr "github.com/nabbar/golib/errors"
	nutsdb "github.com/nutsdb/nutsdb"
)

type NutsDBOptions struct {
	// EntryIdxMode represents using which mode to index the entries.
	EntryIdxMode nutsdb.EntryIdxMode `mapstructure:"entry_idx_mode" json:"entry_idx_mode" yaml:"entry_idx_mode" toml:"entry_idx_mode"`

	// RWMode represents the read and write mode.
	// RWMode includes two options: FileIO and MMap.
	// FileIO represents the read and write mode using standard I/O.
	// MMap represents the read and write mode using mmap.
	RWMode nutsdb.RWMode `mapstructure:"rw_mode" json:"rw_mode" yaml:"rw_mode" toml:"rw_mode"`

	// SegmentSize default value is 8 MBytes
	SegmentSize int64 `mapstructure:"segment_size" json:"segment_size" yaml:"segment_size" toml:"segment_size"`

	// SyncEnable represents if call Sync() function.
	// if SyncEnable is false, high write performance but potential data loss likely.
	// if SyncEnable is true, slower but persistent.
	SyncEnable bool `mapstructure:"sync_enable" json:"sync_enable" yaml:"sync_enable" toml:"sync_enable"`
}

func (o NutsDBOptions) GetNutsDBOptions(dataDir string) nutsdb.Options {
	d := nutsdb.DefaultOptions

	if len(dataDir) < 1 {
		d.RWMode = nutsdb.MMap
	} else {
		d.Dir = dataDir

		//nolint #exhaustive
		switch o.RWMode {
		case nutsdb.MMap:
			d.RWMode = nutsdb.MMap
		default:
			d.RWMode = nutsdb.FileIO
		}
	}

	//nolint #exhaustive
	switch o.EntryIdxMode {
	case nutsdb.HintKeyAndRAMIdxMode:
		d.EntryIdxMode = nutsdb.HintKeyAndRAMIdxMode
	case nutsdb.HintBPTSparseIdxMode:
		d.EntryIdxMode = nutsdb.HintBPTSparseIdxMode
	default:
		d.EntryIdxMode = nutsdb.HintKeyValAndRAMIdxMode
	}

	if o.SegmentSize > 0 {
		d.SegmentSize = o.SegmentSize
	}

	if o.SyncEnable {
		d.SyncEnable = true
	} else {
		d.SyncEnable = false
	}

	return d
}

func (o NutsDBOptions) Validate() liberr.Error {
	err := ErrorValidateConfig.Error(nil)

	if er := libval.New().Struct(o); er != nil {
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
