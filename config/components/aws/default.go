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

package aws

import (
	"bytes"
	"encoding/json"

	libaws "github.com/nabbar/golib/aws"
	cfgstd "github.com/nabbar/golib/aws/configAws"
	cfgcus "github.com/nabbar/golib/aws/configCustom"
	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libsts "github.com/nabbar/golib/status/config"
	spfcbr "github.com/spf13/cobra"
	spfvbr "github.com/spf13/viper"
)

var _defaultConfigStandard = []byte(`{
  "bucket": "",
  "accesskey": "",
  "secretkey": "",
  "region": "",
  "endpoint": ""
}`)

var _defaultConfigStandardWithStatus = []byte(`{
  "bucket": "",
  "accesskey": "",
  "secretkey": "",
  "region": "",
  "endpoint": "",
  "status":` + string(libsts.DefaultConfig(libcfg.JSONIndent+libcfg.JSONIndent)) + `
}`)

var _defaultConfigCustom = []byte(`{
  "bucket": "",
  "accesskey": "",
  "secretkey": "",
  "region": "",
  "endpoint": ""
}`)

var _defaultConfigCustomWithStatus = []byte(`{
  "bucket": "",
  "accesskey": "",
  "secretkey": "",
  "region": "",
  "endpoint": "",
  "status":` + string(libsts.DefaultConfig(libcfg.JSONIndent+libcfg.JSONIndent)) + `
}`)

var _defaultConfig = _defaultConfigCustom

func SetDefaultConfig(cfg []byte) {
	_defaultConfig = cfg
}

func SetDefaultConfigStandard(withStatus bool) {
	if withStatus {
		_defaultConfig = _defaultConfigStandardWithStatus
	} else {
		_defaultConfig = _defaultConfigStandard
	}
}

func SetDefaultConfigCustom(withStatus bool) {
	if withStatus {
		_defaultConfig = _defaultConfigCustomWithStatus
	} else {
		_defaultConfig = _defaultConfigCustom
	}
}

func DefaultConfig(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfig, indent, libcfg.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func (c *componentAws) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}

func (c *componentAws) RegisterFlag(Command *spfcbr.Command, Viper *spfvbr.Viper) error {
	_ = Command.PersistentFlags().String(c.key+".access-key", "", "AWS Access Key")
	_ = Command.PersistentFlags().String(c.key+".secret-key", "", "AWS Secret Key")
	_ = Command.PersistentFlags().String(c.key+".bucket", "", "Bucket to use")
	_ = Command.PersistentFlags().String(c.key+".region", "", "Region for bucket")
	_ = Command.PersistentFlags().String(c.key+".endpoint", "", "Endpoint if necessary for the region")

	if err := Viper.BindPFlag(c.key+".access-key", Command.PersistentFlags().Lookup(c.key+".access-key")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".secret-key", Command.PersistentFlags().Lookup(c.key+".secret-key")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".bucket", Command.PersistentFlags().Lookup(c.key+".bucket")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".region", Command.PersistentFlags().Lookup(c.key+".region")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".endpoint", Command.PersistentFlags().Lookup(c.key+".endpoint")); err != nil {
		return err
	}

	return nil
}

func (c *componentAws) _getConfig(getCfg libcfg.FuncComponentConfigGet) (libaws.Config, *libsts.ConfigStatus, liberr.Error) {
	var (
		vpr = c.vpr()
		cfg libaws.Config
		sts *libsts.ConfigStatus
		err liberr.Error
	)

	switch c.d {
	case ConfigCustomStatus:
		cnf := cfgcus.ModelStatus{}
		if e := getCfg(c.key, &cnf); e != nil {
			return nil, nil, ErrorParamInvalid.Error(e)
		}
		if s := vpr.GetString(c.key + ".access-key"); s != "" {
			cnf.Config.AccessKey = s
		}
		if s := vpr.GetString(c.key + ".secret-key"); s != "" {
			cnf.Config.SecretKey = s
		}
		if s := vpr.GetString(c.key + ".bucket"); s != "" {
			cnf.Config.Bucket = s
		}
		if s := vpr.GetString(c.key + ".region"); s != "" {
			cnf.Config.Region = s
		}
		if s := vpr.GetString(c.key + ".endpoint"); s != "" {
			cnf.Config.Endpoint = s
		}
		if cfg, err = c.d.NewFromModel(cnf); err != nil {
			return nil, nil, err
		} else {
			sts = &cnf.Status
		}
	case ConfigCustom:
		cnf := cfgcus.Model{}
		if e := getCfg(c.key, &cnf); e != nil {
			return nil, nil, ErrorParamInvalid.Error(e)
		}
		if s := vpr.GetString(c.key + ".access-key"); s != "" {
			cnf.AccessKey = s
		}
		if s := vpr.GetString(c.key + ".secret-key"); s != "" {
			cnf.SecretKey = s
		}
		if s := vpr.GetString(c.key + ".bucket"); s != "" {
			cnf.Bucket = s
		}
		if s := vpr.GetString(c.key + ".region"); s != "" {
			cnf.Region = s
		}
		if s := vpr.GetString(c.key + ".endpoint"); s != "" {
			cnf.Endpoint = s
		}
		if cfg, err = c.d.NewFromModel(cnf); err != nil {
			return nil, nil, err
		}
	case ConfigStandardStatus:
		cnf := cfgstd.ModelStatus{}
		if e := getCfg(c.key, &cnf); e != nil {
			return nil, nil, ErrorParamInvalid.Error(e)
		}
		if s := vpr.GetString(c.key + ".access-key"); s != "" {
			cnf.Config.AccessKey = s
		}
		if s := vpr.GetString(c.key + ".secret-key"); s != "" {
			cnf.Config.SecretKey = s
		}
		if s := vpr.GetString(c.key + ".bucket"); s != "" {
			cnf.Config.Bucket = s
		}
		if s := vpr.GetString(c.key + ".region"); s != "" {
			cnf.Config.Region = s
		}
		if cfg, err = c.d.NewFromModel(cnf); err != nil {
			return nil, nil, err
		} else {
			sts = &cnf.Status
		}
	case ConfigStandard:
		cnf := cfgstd.Model{}
		if e := getCfg(c.key, &cnf); e != nil {
			return nil, nil, ErrorParamInvalid.Error(e)
		}
		if s := vpr.GetString(c.key + ".access-key"); s != "" {
			cnf.AccessKey = s
		}
		if s := vpr.GetString(c.key + ".secret-key"); s != "" {
			cnf.SecretKey = s
		}
		if s := vpr.GetString(c.key + ".bucket"); s != "" {
			cnf.Bucket = s
		}
		if s := vpr.GetString(c.key + ".region"); s != "" {
			cnf.Region = s
		}
		if cfg, err = c.d.NewFromModel(cnf); err != nil {
			return nil, nil, err
		}
	}

	if err = cfg.Validate(); err != nil {
		return nil, nil, ErrorConfigInvalid.Error(err)
	}

	return cfg, sts, nil
}
