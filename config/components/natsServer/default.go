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

package natsServer

import (
	"bytes"
	"encoding/json"

	libcfg "github.com/nabbar/golib/config"
	cpttls "github.com/nabbar/golib/config/components/tls"
	liberr "github.com/nabbar/golib/errors"
	libnat "github.com/nabbar/golib/nats"
	spfcbr "github.com/spf13/cobra"
	spfvbr "github.com/spf13/viper"
)

var _defaultConfig = []byte(`{
   "server":{
      "name":"node-0",
      "host":"127.0.0.1",
      "port":9000,
      "client_advertise":"",
      "http_host":"127.0.0.1",
      "http_port":9200,
      "https_port":0,
      "http_base_path":"",
      "prof_port":9300,
      "pid_file":"",
      "ports_file_dir":"",
      "routes":[
         {
            "Scheme":"nats",
            "Opaque":"",
            "User":{
               
            },
            "Host":"127.0.0.1:9101",
            "Path":"",
            "RawPath":"",
            "ForceQuery":false,
            "RawQuery":"",
            "Fragment":"",
            "RawFragment":""
         },
         {
            "Scheme":"nats",
            "Opaque":"",
            "User":{
               
            },
            "Host":"127.0.0.1:9102",
            "Path":"",
            "RawPath":"",
            "ForceQuery":false,
            "RawQuery":"",
            "Fragment":"",
            "RawFragment":""
         },
         {
            "Scheme":"nats",
            "Opaque":"",
            "User":{
               
            },
            "Host":"127.0.0.1:9103",
            "Path":"",
            "RawPath":"",
            "ForceQuery":false,
            "RawQuery":"",
            "Fragment":"",
            "RawFragment":""
         }
      ],
      "routes_str":"nats://127.0.0.1:9101,nats://127.0.0.1:9102,nats://127.0.0.1:9103",
      "no_log":true,
      "username":"",
      "password":"",
      "token":"",
      "jet_stream":true,
      "jet_stream_max_memory":0,
      "jet_stream_max_store":0,
      "store_dir":"/path/to/working/folder",
      "permission_store_dir":"0755",
      "tags":[
         ""
      ],
      "tls":false,
      "allow_no_tls":true,
      "tls_timeout":0,
      "tls_config":` + string(cpttls.DefaultConfig(libcfg.JSONIndent+libcfg.JSONIndent)) + `
   },
   "cluster":{
      "name":"Test-cluster",
      "host":"127.0.0.1",
      "port":9100,
      "listen_str":"",
      "advertise":"",
      "no_advertise":false,
      "connect_retries":5,
      "username":"",
      "password":"",
      "auth_timeout":0,
      "permissions":{
         "import":{
            "allow":null,
            "deny":null
         },
         "export":{
            "allow":null,
            "deny":null
         }
      },
      "tls":false,
      "tls_timeout":0,
      "tls_config":` + string(cpttls.DefaultConfig(libcfg.JSONIndent+libcfg.JSONIndent)) + `
   },
   "gateways":{
      "name":"",
      "host":"",
      "port":0,
      "username":"",
      "password":"",
      "auth_timeout":0,
      "advertise":"",
      "connect_retries":0,
      "gateways":null,
      "reject_unknown":false,
      "tls":false,
      "tls_timeout":0,
      "tls_config":` + string(cpttls.DefaultConfig(libcfg.JSONIndent+libcfg.JSONIndent)) + `
   },
   "leaf":{
      "host":"",
      "port":0,
      "username":"",
      "password":"",
      "auth_timeout":0,
      "advertise":"",
      "no_advertise":false,
      "account":"",
      "users":null,
      "reconnect_interval":0,
      "remotes":null,
      "tls":false,
      "tls_timeout":0,
      "tls_config":` + string(cpttls.DefaultConfig(libcfg.JSONIndent+libcfg.JSONIndent)) + `
   },
   "websockets":{
      "host":"",
      "port":0,
      "advertise":"",
      "no_auth_user":"",
      "jwt_cookie":"",
      "username":"",
      "password":"",
      "token":"",
      "auth_timeout":0,
      "same_origin":false,
      "allowed_origins":null,
      "compression":false,
      "no_tls":false,
      "handshake_timeout":0,
      "tls_config":` + string(cpttls.DefaultConfig(libcfg.JSONIndent+libcfg.JSONIndent)) + `
   },
   "mqtt":{
      "host":"",
      "port":0,
      "no_auth_user":"",
      "username":"",
      "password":"",
      "token":"",
      "auth_timeout":0,
      "ack_wait":0,
      "max_ack_pending":0,
      "tls":false,
      "tls_timeout":0,
      "tls_config": ` + string(cpttls.DefaultConfig(libcfg.JSONIndent+libcfg.JSONIndent)) + `
   },
   "limits":{
      "max_conn":0,
      "max_subs":0,
      "ping_interval":0,
      "max_pings_out":0,
      "max_control_line":0,
      "max_payload":0,
      "max_pending":0,
      "write_deadline":0,
      "max_closed_clients":0,
      "lame_duck_duration":0,
      "lame_duck_grace_period":0,
      "no_sublist_cache":false,
      "no_header_support":false,
      "disable_short_first_ping":false
   },
   "logs":{
      "log_file":"/path/to/log/file.log",
      "permission_folder":"0755",
      "permission_file":"0644",
      "syslog":false,
      "remote_syslog":"",
      "log_size_limit":0,
      "max_traced_msg_len":0,
      "connect_error_reports":0,
      "reconnect_error_reports":0
   },
   "auth":{
      "nkeys":null,
      "users":[
         {
            "username":"username",
            "password":"password",
            "account":"cluster",
            "connection_types":[
               "STANDARD",
               "LEAFNODE",
               "WEBSOCKET",
               "MQTT"
            ]
         }
      ],
      "accounts":[
         {
            "name":"cluster",
            "permission":{
               "publish":{
                  "allow":[
                     ">",
                     "*"
                  ],
                  "deny":[]
               },
               "subscribe":{
                  "allow":[
                     ">",
                     "*"
                  ],
                  "deny":[]
               },
               "response":{
                  "max_msgs":1000000000,
                  "expires":1
               }
            }
         }
      ],
      "auth_timeout":0,
      "no_auth_user":"",
      "system_account":"cluster",
      "no_system_account":false,
      "allow_new_accounts":true,
      "trusted_keys":[],
      "trusted_operators":[]
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

func (c *componentNats) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}

func (c *componentNats) RegisterFlag(Command *spfcbr.Command, Viper *spfvbr.Viper) error {
	return nil
}

func (c *componentNats) _getConfig(getCfg libcfg.FuncComponentConfigGet) (libnat.Config, liberr.Error) {
	var (
		cfg = libnat.Config{}
		err liberr.Error
	)

	if e := getCfg(c.key, &cfg); e != nil {
		return cfg, ErrorParamInvalid.Error(e)
	}

	if err = cfg.Validate(); err != nil {
		return cfg, ErrorConfigInvalid.Error(err)
	}

	return cfg, nil
}
