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
	"net/url"

	libaws "github.com/nabbar/golib/aws"
	cfgstd "github.com/nabbar/golib/aws/configAws"
	cfgcus "github.com/nabbar/golib/aws/configCustom"
	liberr "github.com/nabbar/golib/errors"
)

type ConfigDriver uint8

const (
	ConfigStandard ConfigDriver = iota
	ConfigCustom
)

func DriverConfig(value int) ConfigDriver {
	switch value {
	case int(ConfigCustom):
		return ConfigCustom
	default:
		return ConfigStandard
	}
}

func (a ConfigDriver) String() string {
	switch a {
	case ConfigCustom:
		return "Custom"
	default:
		return "Standard"
	}
}

func (a ConfigDriver) Unmarshal(p []byte) (libaws.Config, liberr.Error) {
	switch a {
	case ConfigCustom:
		return cfgcus.NewConfigJsonUnmashal(p)
	default:
		return cfgstd.NewConfigJsonUnmashal(p)
	}
}

func (a ConfigDriver) Config(bucket, accessKey, secretKey string, region string, endpoint *url.URL) libaws.Config {
	switch a {
	case ConfigCustom:
		return cfgcus.NewConfig(bucket, accessKey, secretKey, endpoint, region)
	default:
		return cfgstd.NewConfig(bucket, accessKey, secretKey, region)
	}
}

func (a ConfigDriver) Model() interface{} {
	switch a {
	case ConfigCustom:
		return cfgcus.Model{}
	default:
		return cfgstd.Model{}
	}
}

func (a ConfigDriver) NewFromModel(i interface{}) (libaws.Config, liberr.Error) {
	switch a {
	case ConfigCustom:
		if o, ok := i.(cfgcus.Model); !ok {
			return nil, ErrorConfigInvalid.Error(nil)
		} else {
			if edp, err := url.Parse(o.Endpoint); err != nil {
				return nil, ErrorConfigInvalid.ErrorParent(err)
			} else {
				cfg := cfgcus.NewConfig(o.Bucket, o.AccessKey, o.SecretKey, edp, o.Region)

				if e := cfg.RegisterRegionAws(edp); e != nil {
					return cfg, e
				}

				return cfg, nil
			}
		}

	default:
		if o, ok := i.(cfgstd.Model); !ok {
			return nil, ErrorConfigInvalid.Error(nil)
		} else {
			return cfgstd.NewConfig(o.Bucket, o.AccessKey, o.SecretKey, o.Region), nil
		}
	}
}
