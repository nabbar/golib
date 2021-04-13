//+build amd64 arm64 arm64be ppc64 ppc64le mips64 mips64le riscv64 s390x sparc64 wasm

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
	"time"

	"github.com/go-playground/validator/v10"
	dgbcfg "github.com/lni/dragonboat/v3/config"
	liberr "github.com/nabbar/golib/errors"
)

type ConfigExpert struct {
	// Engine is the cponfiguration for the execution engine.
	Engine ConfigEngine `mapstructure:"engine" json:"engine" yaml:"engine" toml:"engine"`

	// TestNodeHostID is the NodeHostID value to be used by the NodeHost instance.
	// This field is expected to be used in tests only.
	TestNodeHostID uint64 `mapstructure:"test_node_host_id" json:"test_node_host_id" yaml:"test_node_host_id" toml:"test_node_host_id"`

	// TestGossipProbeInterval define the probe interval used by the gossip
	// service in tests.
	TestGossipProbeInterval time.Duration `mapstructure:"test_gossip_probe_interval" json:"test_gossip_probe_interval" yaml:"test_gossip_probe_interval" toml:"test_gossip_probe_interval"`
}

func (c ConfigExpert) GetDGBConfigExpert() dgbcfg.ExpertConfig {
	d := dgbcfg.ExpertConfig{
		Engine: c.Engine.GetDGBConfigEngine(),
	}

	if c.TestNodeHostID > 0 {
		d.TestNodeHostID = c.TestNodeHostID
	}

	if c.TestGossipProbeInterval > 0 {
		d.TestGossipProbeInterval = c.TestGossipProbeInterval
	}

	return d
}

func (c ConfigExpert) Validate() liberr.Error {
	val := validator.New()
	err := val.Struct(c)

	if e, ok := err.(*validator.InvalidValidationError); ok {
		return ErrorValidateExpert.ErrorParent(e)
	}

	out := ErrorValidateExpert.Error(nil)

	for _, e := range err.(validator.ValidationErrors) {
		//nolint goerr113
		out.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Field(), e.ActualTag()))
	}

	if out.HasParent() {
		return out
	}

	return nil
}
