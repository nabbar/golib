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

	cfgcst "github.com/nabbar/golib/config/const"
	libhtc "github.com/nabbar/golib/httpcli"
	montps "github.com/nabbar/golib/monitor/types"
)

var _defaultConfigStandard = []byte(`{
  "bucket": "",
  "accesskey": "",
  "secretkey": "",
  "region": "",
  "endpoint": ""
}`)

var _defaultConfigStandardWithStatus = []byte(`{
  "config":` + string(DefaultConfigStandard(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `,
  "http-client":` + string(libhtc.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `,
  "health":` + string(montps.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `
}`)

var _defaultConfigCustom = []byte(`{
  "bucket": "",
  "accesskey": "",
  "secretkey": "",
  "region": "",
  "endpoint": ""
}`)

var _defaultConfigCustomWithStatus = []byte(`{
  "config":` + string(DefaultConfigCustom(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `,
  "http-client":` + string(libhtc.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `,
  "health":` + string(montps.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `
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

func DefaultConfigStandard(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfigStandard, indent, cfgcst.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func DefaultConfigStandardStatus(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfigStandardWithStatus, indent, cfgcst.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func DefaultConfigCustom(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfigCustom, indent, cfgcst.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func DefaultConfigCustomStatus(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfigCustomWithStatus, indent, cfgcst.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func DefaultConfig(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfig, indent, cfgcst.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func (o *componentAws) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}
