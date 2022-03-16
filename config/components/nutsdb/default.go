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

package nutsdb

import (
	"bytes"
	"encoding/json"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libndb "github.com/nabbar/golib/nutsdb"
	spfcbr "github.com/spf13/cobra"
	spfvbr "github.com/spf13/viper"
)

var _defaultConfig = []byte(`{
   "db":{
      "entry_idx_mode":0,
      "rw_mode":0,
      "segment_size":8388608,
      "sync_enable":false,
      "start_file_loading_mode":1
   },
   "cluster":{
      "node":{
         "deployment_id":0,
         "wal_dir":"",
         "node_host_dir":"",
         "rtt_millisecond":200,
         "raft_address":"localhost:9001",
         "address_by_node_host_id":false,
         "listen_address":"",
         "mutual_tls":false,
         "ca_file":"",
         "cert_file":"",
         "key_file":"",
         "enable_metrics":true,
         "max_send_queue_size":0,
         "max_receive_queue_size":0,
         "max_snapshot_send_bytes_per_second":0,
         "max_snapshot_recv_bytes_per_second":0,
         "notify_commit":false,
         "gossip":{
            "bind_address":"",
            "advertise_address":"",
            "seed":null
         },
         "expert":{
            "engine":{
               "exec_shards":0,
               "commit_shards":0,
               "apply_shards":0,
               "snapshot_shards":0,
               "close_shards":0
            },
            "test_node_host_id":0,
            "test_gossip_probe_interval":0
         }
      },
      "cluster":{
         "node_id":1,
         "cluster_id":1,
         "check_quorum":true,
         "election_rtt":15,
         "heartbeat_rtt":1,
         "snapshot_entries":10,
         "compaction_overhead":0,
         "ordered_config_change":false,
         "max_in_mem_log_size":0,
         "snapshot_compression":0,
         "entry_compression":0,
         "disable_auto_compactions":true,
         "is_observer":false,
         "is_witness":false,
         "quiesce":false
      },
      "init_member":{
         "1":"localhost:9001"
      }
   },
   "directories":{
      "base":"/tmp/nutsdb/node-%d",
      "sub_data":"data",
      "sub_backup":"backup",
      "sub_temp":"temp",
      "wal_dir":"",
      "host_dir":"",
      "limit_number_backup":5,
      "permission":504
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

func (c *componentNutsDB) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}

func (c *componentNutsDB) RegisterFlag(Command *spfcbr.Command, Viper *spfvbr.Viper) error {
	return nil
}

func (c *componentNutsDB) _getConfig(getCfg libcfg.FuncComponentConfigGet) (libndb.Config, liberr.Error) {
	var (
		cfg = libndb.Config{}
		err liberr.Error
	)

	if e := getCfg(c.key, &cfg); e != nil {
		return cfg, ErrorParamsInvalid.Error(e)
	}

	if err = cfg.Validate(); err != nil {
		return cfg, ErrorConfigInvalid.Error(err)
	}

	return cfg, nil
}
