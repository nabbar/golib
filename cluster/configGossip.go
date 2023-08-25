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
	dgbcfg "github.com/lni/dragonboat/v3/config"
	liberr "github.com/nabbar/golib/errors"
)

type ConfigGossip struct {
	// BindAddress is the address for the gossip service to bind to and listen on.
	// Both UDP and TCP ports are used by the gossip service. The local gossip
	// service should be able to receive gossip service related messages by
	// binding to and listening on this address. BindAddress is usually in the
	// format of IP:Port, Hostname:Port or DNS Name:Port.
	BindAddress string `mapstructure:"bind_address" json:"bind_address" yaml:"bind_address" toml:"bind_address" validate:"omitempty,hostname_port"`

	// AdvertiseAddress is the address to advertise to other NodeHost instances
	// used for NAT traversal. Gossip services running on remote NodeHost
	// instances will use AdvertiseAddress to exchange gossip service related
	// messages. AdvertiseAddress is in the format of IP:Port.
	AdvertiseAddress string `mapstructure:"advertise_address" json:"advertise_address" yaml:"advertise_address" toml:"advertise_address" validate:"omitempty,printascii"`

	// Seed is a list of AdvertiseAddress of remote NodeHost instances. Local
	// NodeHost instance will try to contact all of them to bootstrap the gossip
	// service. At least one reachable NodeHost instance is required to
	// successfully bootstrap the gossip service. Each seed address is in the
	// format of IP:Port, Hostname:Port or DNS Name:Port.
	//
	// It is ok to include seed addresses that are temporarily unreachable, e.g.
	// when launching the first NodeHost instance in your deployment, you can
	// include AdvertiseAddresses from other NodeHost instances that you plan to
	// launch shortly afterwards.
	Seed []string `mapstructure:"seed" json:"seed" yaml:"seed" toml:"seed"`
}

func (c ConfigGossip) GetDGBConfigGossip() dgbcfg.GossipConfig {
	d := dgbcfg.GossipConfig{}

	if c.BindAddress != "" {
		d.BindAddress = c.BindAddress
	}

	if c.AdvertiseAddress != "" {
		d.AdvertiseAddress = c.AdvertiseAddress
	}

	if len(c.Seed) > 0 {
		d.Seed = make([]string, 0)

		for _, v := range c.Seed {

			if v == "" {
				continue
			}

			d.Seed = append(d.Seed, v)
		}
	}

	return d
}

func (c ConfigGossip) Validate() liberr.Error {
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
