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

package tls

import (
	"bytes"
	"encoding/json"

	cfgtps "github.com/nabbar/golib/config/const"
)

var _defaultConfig = []byte(`{
   "inheritDefault": true,
   "versionMin":"1.2",
   "versionMax":"1.2",
   "dynamicSizingDisable":false,
   "sessionTicketDisable":false,
   "authClient":"none",
   "curveList":[
      "X25519",
      "P256",
      "P384",
      "P521"
   ],
   "cipherList":[
      "RSA-AES128-GCM",
      "RSA-AES128-CBC",
      "RSA-AES256-GCM",
      "RSA-CHACHA",
      "ECDSA-AES128-GCM",
      "ECDSA-AES128-CBC",
      "ECDSA-AES256-GCM",
      "ECDSA-CHACHA",
      "ECDHE-RSA-AES128-GCM",
      "ECDHE-RSA-AES128-CBC",
      "ECDHE-RSA-AES256-GCM",
      "ECDHE-RSA-CHACHA",
      "ECDHE-ECDSA-AES128-GCM",
      "ECDHE-ECDSA-AES128-CBC",
      "ECDHE-ECDSA-AES256-GCM",
      "ECDHE-ECDSA-CHACHA"
   ],
   "rootCA":[
      ""
   ],
   "rootCAFiles":[
      ""
   ],
   "clientCA":[
      ""
   ],
   "clientCAFiles":[
      ""
   ],
   "certPair":[
      {
         "key":"",
         "pem":""
      }
   ],
   "certPairFiles":[
      {
         "key":"",
         "pem":""
      }
   ]
}`)

func SetDefaultConfig(cfg []byte) {
	_defaultConfig = cfg
}

func DefaultConfig(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfig, indent, cfgtps.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func (o *componentTls) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}
