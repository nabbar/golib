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

	cfgcst "github.com/nabbar/golib/config/const"
)

var _defaultConfig = []byte(`
{
   "disableStandard":false,
   "disableStack":false,
   "disableTimestamp":false,
   "enableTrace":true,
   "traceFilter":"",
   "disableColor":false,
   "logFile":[
      {
         "logLevel":[
            "Debug",
            "Info",
            "Warning",
            "Error",
            "Fatal",
            "Critical"
         ],
         "filepath":"",
         "create":false,
         "createPath":false,
         "fileMode":"0644",
         "pathMode":"0755",
         "disableStack":false,
         "disableTimestamp":false,
         "enableTrace":true
      }
   ],
   "logSyslog":[
      {
         "logLevel":[
            "Debug",
            "Info",
            "Warning",
            "Error",
            "Fatal",
            "Critical"
         ],
         "network":"tcp",
         "host":"",
         "severity":"Error",
         "facility":"local0",
         "tag":"",
         "disableStack":false,
         "disableTimestamp":false,
         "enableTrace":true
      }
   ]
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
