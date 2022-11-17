/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package config

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/nabbar/golib/status"

	libcfg "github.com/nabbar/golib/config"
	libsts "github.com/nabbar/golib/status"
	spfcbr "github.com/spf13/cobra"
	spfvpr "github.com/spf13/viper"
)

type ConfigStatus struct {
	// Mandatory define if the component must be available for the api.
	// If yes, api status will be KO if this component is down.
	// If no, api status can be OK if this component is down.
	Mandatory bool `json:"mandatory" yaml:"mandatory" toml:"mandatory" mapstructure:"mandatory"`

	// MessageOK define the message if the status is OK. Default is "OK"
	MessageOK string `json:"message_ok" yaml:"message_ok" toml:"message_ok" mapstructure:"message_ok" validate:"printascii"`

	// MessageKO define the message if the status is KO. Default is "KO"
	MessageKO string `json:"message_ko" yaml:"message_ko" toml:"message_ko" mapstructure:"message_ko" validate:"printascii"`

	// CacheTimeoutInfo define the time between checking the component information (name, release, ...), to prevent asking it too many. Default is 1 hour.
	CacheTimeoutInfo time.Duration `json:"cache_timeout_info" yaml:"cache_timeout_info" toml:"cache_timeout_info" mapstructure:"cache_timeout_info"`

	// CacheTimeoutHealth define the time between checking the component health to prevent asking it too many. Default is 5 second.
	CacheTimeoutHealth time.Duration `json:"cache_timeout_health" yaml:"cache_timeout_health" toml:"cache_timeout_health" mapstructure:"cache_timeout_health"`
}

var _defaultConfig = []byte(`{
  "mandatory": false,
  "message_ok": "OK",
  "message_ko": "KO",
  "cache_timeout_info": "30s",
  "cache_timeout_health": "5s"
}
`)

func DefaultConfig(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfig, indent, libcfg.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func RegisterFlag(prefix string, Command *spfcbr.Command, Viper *spfvpr.Viper) error {
	_ = Command.PersistentFlags().Bool(prefix+".mandatory", true, "define if the component must be available for the api. If yes, api status will be KO if this component is down. If no, api status can be OK if this component is down.")
	_ = Command.PersistentFlags().String(prefix+".message_ok", status.DefMessageOK, "define the message if the status is OK.")
	_ = Command.PersistentFlags().String(prefix+".message_ko", status.DefMessageKO, "define the message if the status is KO.")
	_ = Command.PersistentFlags().Duration(prefix+".cache_timeout_info", time.Hour, "define the time between checking the component information (name, release, ...), to prevent asking it too many.")
	_ = Command.PersistentFlags().Duration(prefix+".cache_timeout_health", 5*time.Second, "define the time between checking the component health to prevent asking it too many.")

	if err := Viper.BindPFlag(prefix+".mandatory", Command.PersistentFlags().Lookup(prefix+".mandatory")); err != nil {
		return err
	} else if err = Viper.BindPFlag(prefix+".message_ok", Command.PersistentFlags().Lookup(prefix+".message_ok")); err != nil {
		return err
	} else if err = Viper.BindPFlag(prefix+".message_ko", Command.PersistentFlags().Lookup(prefix+".message_ko")); err != nil {
		return err
	} else if err = Viper.BindPFlag(prefix+".cache_timeout_info", Command.PersistentFlags().Lookup(prefix+".cache_timeout_info")); err != nil {
		return err
	} else if err = Viper.BindPFlag(prefix+".cache_timeout_health", Command.PersistentFlags().Lookup(prefix+".cache_timeout_health")); err != nil {
		return err
	}

	return nil
}

func (c *ConfigStatus) fctMessage() (msgOk string, msgKO string) {
	if c.MessageOK == "" {
		c.MessageOK = status.DefMessageOK
	}

	if c.MessageKO == "" {
		c.MessageKO = status.DefMessageKO
	}

	return c.MessageOK, c.MessageKO
}

func (c *ConfigStatus) RegisterStatus(sts libsts.RouteStatus, key string, fctInfo libsts.FctInfo, fctHealth libsts.FctHealth) {
	sts.ComponentNew(key, status.NewComponent(key, c.Mandatory, fctInfo, fctHealth, c.fctMessage, c.CacheTimeoutInfo, c.CacheTimeoutHealth))
}

func (c *ConfigStatus) Component(key string, fctInfo libsts.FctInfo, fctHealth libsts.FctHealth) libsts.Component {
	return status.NewComponent(key, c.Mandatory, fctInfo, fctHealth, c.fctMessage, c.CacheTimeoutInfo, c.CacheTimeoutHealth)
}
