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
	libaws "github.com/nabbar/golib/aws"
	cfgstd "github.com/nabbar/golib/aws/configAws"
	cfgcus "github.com/nabbar/golib/aws/configCustom"
	libhtc "github.com/nabbar/golib/httpcli"
	libreq "github.com/nabbar/golib/request"
	spfcbr "github.com/spf13/cobra"
	spfvpr "github.com/spf13/viper"
)

type _configFlag struct {
	accessKey *string
	secretKey *string
	bucket    *string
	region    *string
	endpoint  *string
}

func (c *_configFlag) updCustom(cfg *cfgcus.Model) {
	if c.accessKey != nil {
		cfg.AccessKey = *c.accessKey
	}
	if c.secretKey != nil {
		cfg.SecretKey = *c.secretKey
	}
	if c.bucket != nil {
		cfg.Bucket = *c.bucket
	}
	if c.region != nil {
		cfg.Region = *c.region
	}
	if c.endpoint != nil {
		cfg.Endpoint = *c.endpoint
	}
}

func (c *_configFlag) updStandard(cfg *cfgstd.Model) {
	if c.accessKey != nil {
		cfg.AccessKey = *c.accessKey
	}
	if c.secretKey != nil {
		cfg.SecretKey = *c.secretKey
	}
	if c.bucket != nil {
		cfg.Bucket = *c.bucket
	}
	if c.region != nil {
		cfg.Region = *c.region
	}
}

func (o *componentAws) RegisterFlag(Command *spfcbr.Command) error {
	var (
		key = o._getKey()
		vpr *spfvpr.Viper
	)

	if vpr = o._getSPFViper(); vpr == nil {
		return ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return ErrorComponentNotInitialized.Error(nil)
	}

	_ = Command.PersistentFlags().String(key+".access-key", "", "AWS Access Key")
	_ = Command.PersistentFlags().String(key+".secret-key", "", "AWS Secret Key")
	_ = Command.PersistentFlags().String(key+".bucket", "", "Bucket to use")
	_ = Command.PersistentFlags().String(key+".region", "", "Region for bucket")
	_ = Command.PersistentFlags().String(key+".endpoint", "", "Endpoint if necessary for the region")

	if err := vpr.BindPFlag(key+".access-key", Command.PersistentFlags().Lookup(key+".access-key")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".secret-key", Command.PersistentFlags().Lookup(key+".secret-key")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".bucket", Command.PersistentFlags().Lookup(key+".bucket")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".region", Command.PersistentFlags().Lookup(key+".region")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".endpoint", Command.PersistentFlags().Lookup(key+".endpoint")); err != nil {
		return err
	}

	return nil
}

func (o *componentAws) _getConfig() (libaws.Config, *libreq.OptionsHealth, *libhtc.Options, error) {
	var (
		key string
		cfg libaws.Config
		flg = o._getFlagUpdate()
		mon *libreq.OptionsHealth
		htc *libhtc.Options
		vpr *spfvpr.Viper
		err error
	)

	if vpr = o._getSPFViper(); vpr == nil {
		return nil, nil, nil, ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return nil, nil, nil, ErrorComponentNotInitialized.Error(nil)
	}

	switch o.d {
	case ConfigCustomStatus:
		cnf := cfgcus.ModelStatus{}
		if err = vpr.UnmarshalKey(key, &cnf); err != nil {
			return nil, nil, nil, ErrorParamInvalid.Error(err)
		} else {
			flg.updCustom(&cnf.Config)
		}

		if cfg, err = o.d.NewFromModel(cnf); err != nil {
			return nil, nil, nil, err
		} else {
			mon = &cnf.Monitor
			htc = &cnf.HTTPClient
		}

	case ConfigCustom:
		cnf := cfgcus.Model{}
		if err = vpr.UnmarshalKey(key, &cnf); err != nil {
			return nil, nil, nil, ErrorParamInvalid.Error(err)
		} else {
			flg.updCustom(&cnf)
		}

		if cfg, err = o.d.NewFromModel(cnf); err != nil {
			return nil, nil, nil, err
		} else {
			mon = nil
			htc = nil
		}

	case ConfigStandardStatus:
		cnf := cfgstd.ModelStatus{}
		if err = vpr.UnmarshalKey(key, &cnf); err != nil {
			return nil, nil, nil, ErrorParamInvalid.Error(err)
		} else {
			flg.updStandard(&cnf.Config)
		}

		if cfg, err = o.d.NewFromModel(cnf); err != nil {
			return nil, nil, nil, err
		} else {
			mon = &cnf.Monitor
			htc = &cnf.HTTPClient
		}

	case ConfigStandard:
		cnf := cfgstd.Model{}
		if err = vpr.UnmarshalKey(key, &cnf); err != nil {
			return nil, nil, nil, ErrorParamInvalid.Error(err)
		} else {
			flg.updStandard(&cnf)
		}

		if cfg, err = o.d.NewFromModel(cnf); err != nil {
			return nil, nil, nil, err
		} else {
			mon = nil
			htc = nil
		}
	}

	if err = cfg.Validate(); err != nil {
		return nil, nil, nil, ErrorConfigInvalid.Error(err)
	}

	if mon != nil {
		if err = mon.Validate(); err != nil {
			return nil, nil, nil, ErrorConfigInvalid.Error(err)
		}
	}

	if htc != nil {
		if err = htc.Validate(); err != nil {
			return nil, nil, nil, ErrorConfigInvalid.Error(err)
		}
	}

	return cfg, mon, htc, nil
}

func (o *componentAws) _getFlagUpdate() *_configFlag {
	var (
		cfg = &_configFlag{
			accessKey: nil,
			secretKey: nil,
			bucket:    nil,
			region:    nil,
			endpoint:  nil,
		}
		vpr *spfvpr.Viper
		key string
	)

	if vpr = o._getSPFViper(); vpr == nil {
		return nil
	} else if key = o._getKey(); len(key) < 1 {
		return nil
	}

	if s := vpr.GetString(key + ".access-key"); s != "" {
		cfg.accessKey = &s
	}
	if s := vpr.GetString(key + ".secret-key"); s != "" {
		cfg.secretKey = &s
	}
	if s := vpr.GetString(key + ".bucket"); s != "" {
		cfg.bucket = &s
	}
	if s := vpr.GetString(key + ".region"); s != "" {
		cfg.region = &s
	}
	if s := vpr.GetString(key + ".endpoint"); s != "" {
		cfg.endpoint = &s
	}

	return cfg
}
