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

	libval "github.com/go-playground/validator/v10"
	dgbclt "github.com/lni/dragonboat/v3"
	dgbcfg "github.com/lni/dragonboat/v3/config"
	liberr "github.com/nabbar/golib/errors"
)

type Config struct {
	Node       ConfigNode        `mapstructure:"node" json:"node" yaml:"node" toml:"node" validate:"dive"`
	Cluster    ConfigCluster     `mapstructure:"cluster" json:"cluster" yaml:"cluster" toml:"cluster" validate:"dive"`
	InitMember map[uint64]string `mapstructure:"init_member" json:"init_member" yaml:"init_member" toml:"init_member"`
}

func (c Config) GetDGBConfigCluster() dgbcfg.Config {
	return c.Cluster.GetDGBConfigCluster()
}

func (c Config) GetDGBConfigNode() dgbcfg.NodeHostConfig {
	return c.Node.GetDGBConfigNodeHost()
}

func (c Config) GetInitMember() map[uint64]dgbclt.Target {
	var m = make(map[uint64]dgbclt.Target)

	for k, v := range c.InitMember {
		m[k] = v
	}

	return m
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
