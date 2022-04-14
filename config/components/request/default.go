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

package request

import (
	"bytes"
	"encoding/json"

	libcfg "github.com/nabbar/golib/config"
	cmptls "github.com/nabbar/golib/config/components/tls"
	libsts "github.com/nabbar/golib/status"
	spfcbr "github.com/spf13/cobra"
	spfvbr "github.com/spf13/viper"
)

var _defaultConfig = []byte(`{
   "endpoint":"https://endpoint.example.com/path",
   "http_client": {
     "timeout":"0s",
     "http2": true,
     "tls": ` + string(cmptls.DefaultConfig("  ")) + `,
     "force_ip": {
       "enable": false,
       "net":"tcp",
       "ip":"127.0.0.1:8080"
     }
   },
   "auth": {
     "basic":{
       "enable": false,
       "username":"user",
       "password":"pass"
     },
     "bearer":{
       "enable": false,
       "token":"ws5f4sg5f1ds251cs51cs5dc1sd35c1sd35cv1sd"
     }
   },
   "health": {
     "enable": false,
     "endpoint":"https://endpoint.example.com/healcheck",
     "auth": {
       "basic":{
         "enable": false,
         "username":"user",
         "password":"pass"
       },
       "bearer":{
         "enable": false,
         "token":"ws5f4sg5f1ds251cs51cs5dc1sd35c1sd35cv1sd"
       }
     },
     "result": {
       "valid_http_code": [200, 201, 202, 203, 204],
       "invalid_http_code": [401, 403, 404, 405, 500, 501, 502, 503, 504],
       "contain": ["OK", "Done"],
       "not_contain": ["KO", "fail", "error"],
     },
     "status": ` + string(libsts.DefaultConfig("  ")) + `
   }
}`)

func SetDefaultConfig(cfg []byte) {
	_defaultConfig = cfg
}

func DefaultConfig(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfig, indent, libcfg.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func (c *componentRequest) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}

func (c *componentRequest) RegisterFlag(Command *spfcbr.Command, Viper *spfvbr.Viper) error {
	return nil
}
