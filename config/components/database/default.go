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

package database

import (
	"bytes"
	"encoding/json"

	cfgcst "github.com/nabbar/golib/config/const"
	moncfg "github.com/nabbar/golib/monitor/types"
)

var _defaultConfig = []byte(`{
  "driver": "",
  "name": "",
  "dsn": "",
  "skip-default-transaction": false,
  "full-save-associations": false,
  "dry-run": false,
  "prepare-stmt": false,
  "disable-automatic-ping": false,
  "disable-foreign-key-constraint-when-migrating": false,
  "disable-nested-transaction": false,
  "allow-global-update": false,
  "query-fields": false,
  "create-batch-size": 0,
  "enable-connection-pool": false,
  "pool-max-idle-conns": 0,
  "pool-max-open-conns": 0,
  "pool-conn-max-lifetime": "0s",
  "disabled": false,
  "monitor": ` + string(moncfg.DefaultConfig(cfgcst.JSONIndent)) + `
}`)

func SetDefaultConfig(cfg []byte) {
	_defaultConfig = cfg
}

func DefaultConfig(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfig, indent, cfgcst.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func (o *componentDatabase) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}
