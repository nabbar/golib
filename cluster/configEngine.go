//go:build amd64 || arm64 || arm64be || ppc64 || ppc64le || mips64 || mips64le || riscv64 || s390x || sparc64 || wasm
// +build amd64 arm64 arm64be ppc64 ppc64le mips64 mips64le riscv64 s390x sparc64 wasm

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

package cluster

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	dgbcfg "github.com/lni/dragonboat/v3/config"
	liberr "github.com/nabbar/golib/errors"
)

type ConfigEngine struct {
	// ExecShards is the number of execution shards in the first stage of the
	// execution engine. Default value is 16. Once deployed, this value can not
	// be changed later.
	ExecShards uint64 `mapstructure:"exec_shards" json:"exec_shards" yaml:"exec_shards" toml:"exec_shards"`

	// CommitShards is the number of commit shards in the second stage of the
	// execution engine. Default value is 16.
	CommitShards uint64 `mapstructure:"commit_shards" json:"commit_shards" yaml:"commit_shards" toml:"commit_shards"`

	// ApplyShards is the number of apply shards in the third stage of the
	// execution engine. Default value is 16.
	ApplyShards uint64 `mapstructure:"apply_shards" json:"apply_shards" yaml:"apply_shards" toml:"apply_shards"`

	// SnapshotShards is the number of snapshot shards in the forth stage of the
	// execution engine. Default value is 48.
	SnapshotShards uint64 `mapstructure:"snapshot_shards" json:"snapshot_shards" yaml:"snapshot_shards" toml:"snapshot_shards"`

	// CloseShards is the number of close shards used for closing stopped
	// state machines. Default value is 32.
	CloseShards uint64 `mapstructure:"close_shards" json:"close_shards" yaml:"close_shards" toml:"close_shards"`
}

func (c ConfigEngine) GetDGBConfigEngine() dgbcfg.EngineConfig {
	d := dgbcfg.EngineConfig{}

	if c.ExecShards > 0 {
		d.ExecShards = c.ExecShards
	}

	if c.CommitShards > 0 {
		d.CommitShards = c.CommitShards
	}

	if c.ApplyShards > 0 {
		d.ApplyShards = c.ApplyShards
	}

	if c.SnapshotShards > 0 {
		d.SnapshotShards = c.SnapshotShards
	}

	if c.CloseShards > 0 {
		d.CloseShards = c.CloseShards
	}

	return d
}

func (c ConfigEngine) Validate() liberr.Error {
	val := validator.New()
	err := val.Struct(c)

	if e, ok := err.(*validator.InvalidValidationError); ok {
		return ErrorValidateEngine.ErrorParent(e)
	}

	out := ErrorValidateEngine.Error(nil)

	for _, e := range err.(validator.ValidationErrors) {
		//nolint goerr113
		out.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Field(), e.ActualTag()))
	}

	if out.HasParent() {
		return out
	}

	return nil
}
