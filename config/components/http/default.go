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

var _defaultConfig = []byte(`[
   {
      "disabled":false,
      "mandatory":true,
      "timeout_cache_info":"30s",
      "timeout_cache_health":"30s",
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
      "name":"status_http",
      "handler_keys":"status",
      "tls_mandatory":false,
      "listen":"0.0.0.0:6080",
      "expose":"http://0.0.0.0",
      "tls":{
         
      }
   },
   {
      "disabled":false,
      "mandatory":true,
      "timeout_cache_info":"30s",
      "timeout_cache_health":"30s",
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
      "handler_keys":"api",
      "tls_mandatory":false,
      "name":"api_http",
      "listen":"0.0.0.0:7080",
      "expose":"http://0.0.0.0",
      "tls":{
         
      }
   },
   {
      "disabled":false,
      "mandatory":true,
      "timeout_cache_info":"30s",
      "timeout_cache_health":"30s",
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
      "handler_keys":"metrics",
      "tls_mandatory":false,
      "name":"metrics_http",
      "listen":"0.0.0.0:8080",
      "expose":"http://0.0.0.0",
      "tls":{
         
      }
   }
]`)

func (c *componentHttp) DefaultConfig() []byte {
	return _defaultConfig
}

func SetDefaultConfig(cfg []byte) {
	_defaultConfig = cfg
}
