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

package http

import (
	"bytes"
	"encoding/json"

	cpttls "github.com/nabbar/golib/config/components/tls"
	cfgcst "github.com/nabbar/golib/config/const"
	cptlog "github.com/nabbar/golib/logger/config"
	moncfg "github.com/nabbar/golib/monitor/types"
)

var _defaultConfig = []byte(`[
   {
      "disabled":false,
      "name":"status_http",
      "handler_key":"status",
      "listen":"0.0.0.0:6080",
      "expose":"http://0.0.0.0",
      "monitor":` + string(moncfg.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `,
      "read_timeout":"0s",
      "read_header_timeout":"0s",
      "write_timeout":"0s",
      "idle_timeout":"0s",
      "max_header_bytes":0,
      "max_handlers":0,
      "max_concurrent_streams":0,
      "max_read_frame_size":0,
      "permit_prohibited_cipher_suites":false,
      "max_upload_buffer_per_connection":0,
      "max_upload_buffer_per_stream":0,
      "tls_mandatory":false,
      "tls":` + string(cpttls.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `,
      "logger":` + string(cptlog.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `
   },
   {
      "disabled":false,
      "name":"api_http",
      "handler_key":"api",
      "listen":"0.0.0.0:7080",
      "expose":"http://0.0.0.0",
      "monitor":` + string(moncfg.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `,
      "read_timeout":"0s",
      "read_header_timeout":"0s",
      "write_timeout":"0s",
      "idle_timeout":"0s",
      "max_header_bytes":0,
      "max_handlers":0,
      "max_concurrent_streams":0,
      "max_read_frame_size":0,
      "permit_prohibited_cipher_suites":false,
      "max_upload_buffer_per_connection":0,
      "max_upload_buffer_per_stream":0,
      "tls_mandatory":false,
      "tls":` + string(cpttls.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `,
      "logger":` + string(cptlog.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `
   },
   {
      "disabled":false,
      "name":"metrics_http",
      "handler_key":"metrics",
      "listen":"0.0.0.0:8080",
      "expose":"http://0.0.0.0",
      "monitor":` + string(moncfg.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `,
      "read_timeout":"0s",
      "read_header_timeout":"0s",
      "write_timeout":"0s",
      "idle_timeout":"0s",
      "max_header_bytes":0,
      "max_handlers":0,
      "max_concurrent_streams":0,
      "max_read_frame_size":0,
      "permit_prohibited_cipher_suites":false,
      "max_upload_buffer_per_connection":0,
      "max_upload_buffer_per_stream":0,
      "tls_mandatory":false,
      "tls":` + string(cpttls.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `,
      "logger":` + string(cptlog.DefaultConfig(cfgcst.JSONIndent+cfgcst.JSONIndent)) + `
   }
]`)

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

func (o *componentHttp) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}
