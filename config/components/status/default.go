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

package status

import (
	"bytes"
	"encoding/json"
)

// _defaultConfig stores the default JSON configuration for the status component.
var _defaultConfig = []byte(`{
  "return-code": {
    "OK": 200,
    "Warn": 207,
    "KO": 500
  },
  "info": {
    "doc": "http://example.com"
  },
  "component": [
    {
      "mode": "Must",
      "keys": [
        "component1",
        "component2"
      ],
      "configKeys": []
    }
  ]
}`)

// SetDefaultConfig updates the default configuration used by the component.
// This allows applications to provide their own default settings.
func SetDefaultConfig(cfg []byte) {
	_defaultConfig = cfg
}

// DefaultConfig returns the default configuration as a byte slice.
// If an indent string is provided, the JSON output will be formatted accordingly.
func DefaultConfig(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if indent == "" {
		return _defaultConfig
	}
	if err := json.Indent(res, _defaultConfig, "", indent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

// DefaultConfig implements the component interface.
// It returns the default configuration for the status component.
func (o *mod) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}
