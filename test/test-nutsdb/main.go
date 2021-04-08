/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package main

import (
	"context"
	"fmt"
	"time"

	libsem "github.com/nabbar/golib/semaphore"

	liberr "github.com/nabbar/golib/errors"

	libclu "github.com/nabbar/golib/cluster"
	libndb "github.com/nabbar/golib/nutsdb"
	"github.com/xujiajun/nutsdb"
)

const (
	BaseDirPattern = "/nutsdb/node-%d"
	NbInstances    = 3
)

func init() {
	liberr.SetModeReturnError(liberr.ErrorReturnCodeErrorTraceFull)
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())

	defer func() {
		if cnl != nil {
			cnl()
		}
	}()

	cluster := Start(ctx)
	clilst := make([]libndb.Client, NbInstances)

	for i := 0; i < NbInstances; i++ {
		clilst[i] = cluster[i].Client()
	}

	s := libsem.NewSemaphoreWithContext(ctx, 0)
	defer s.DeferMain()

	for i := 0; i < 100; i++ {
		if e := s.NewWorker(); e != nil {
			continue
		}

		go func(sem libsem.Sem, cli libndb.Client, num int) {
			defer sem.DeferWorker()
			Put(cli, fmt.Sprintf("key-%3d", num), fmt.Sprintf("val-%3d", num))
		}(s, clilst[i%3], i)
	}

	time.Sleep(5 * time.Second)

	for i := 0; i < 100; i++ {
		if e := s.NewWorker(); e != nil {
			continue
		}

		c := i%3 + 1
		if c == 3 {
			c = 0
		}

		go func(sem libsem.Sem, cli libndb.Client, num int) {
			defer sem.DeferWorker()
			Get(cli, fmt.Sprintf("key-%3d", num))
		}(s, clilst[c], i)
	}

	if e := s.WaitAll(); e != nil {
		panic(e)
	}
}

func Put(c libndb.Client, key, val string) {
	fmt.Printf("Cmd Put(%s, %s) : %v", key, val, c.Put("myBucket", []byte(key), []byte(val), 0))
}

func Get(c libndb.Client, key string) {
	v, e := c.Get("myBucket", []byte(key))
	fmt.Printf("Cmd Get(%s) : %v --- err : %v", key, v, e)
}

func Start(ctx context.Context) []libndb.NutsDB {
	var clusters = make([]libndb.NutsDB, NbInstances)

	s := libsem.NewSemaphoreWithContext(ctx, 0)
	defer s.DeferMain()

	for i := 0; i < NbInstances; i++ {
		if err := s.NewWorker(); err != nil {
			panic(err)
		}

		clusters[i] = initNutDB(i + 1)

		go func(clu libndb.NutsDB, sem libsem.Sem) {
			defer sem.DeferWorker()
			if err := clu.Listen(); err != nil {
				panic(err)
			}

		}(clusters[i], s)
	}

	if err := s.WaitAll(); err != nil {
		panic(err)
	}

	return clusters
}

func initNutDB(num int) libndb.NutsDB {
	cfg := configNutDB()
	cfg.Cluster.Cluster.NodeID = uint64(num)
	cfg.Cluster.Node.RaftAddress = cfg.Cluster.InitMember[uint64(num)]
	cfg.Directory.Base = fmt.Sprintf(BaseDirPattern, num)

	return libndb.New(cfg)
}

func configNutDB() libndb.Config {
	cfg := libndb.Config{
		DB: libndb.NutsDBOptions{
			EntryIdxMode:         nutsdb.HintKeyValAndRAMIdxMode,
			RWMode:               nutsdb.FileIO,
			SegmentSize:          8 * 1024 * 1024,
			SyncEnable:           false,
			StartFileLoadingMode: nutsdb.MMap,
		},

		Cluster: libclu.Config{
			Node: libclu.ConfigNode{
				DeploymentID:                  0,
				WALDir:                        "",
				NodeHostDir:                   "",
				RTTMillisecond:                1,
				RaftAddress:                   "",
				AddressByNodeHostID:           false,
				ListenAddress:                 "",
				MutualTLS:                     false,
				CAFile:                        "",
				CertFile:                      "",
				KeyFile:                       "",
				EnableMetrics:                 false,
				MaxSendQueueSize:              0,
				MaxReceiveQueueSize:           0,
				MaxSnapshotSendBytesPerSecond: 0,
				MaxSnapshotRecvBytesPerSecond: 0,
				NotifyCommit:                  false,
				Gossip: libclu.ConfigGossip{
					BindAddress:      "",
					AdvertiseAddress: "",
					Seed:             nil,
				},
				Expert: libclu.ConfigExpert{
					Engine: libclu.ConfigEngine{
						ExecShards:     0,
						CommitShards:   0,
						ApplyShards:    0,
						SnapshotShards: 0,
						CloseShards:    0,
					},
					TestNodeHostID:          0,
					TestGossipProbeInterval: 0,
				},
			},
			Cluster: libclu.ConfigCluster{
				NodeID:                  0,
				ClusterID:               1,
				CheckQuorum:             true,
				ElectionRTT:             0,
				HeartbeatRTT:            0,
				SnapshotEntries:         0,
				CompactionOverhead:      0,
				OrderedConfigChange:     false,
				MaxInMemLogSize:         0,
				SnapshotCompressionType: 0,
				EntryCompressionType:    0,
				DisableAutoCompactions:  false,
				IsObserver:              false,
				IsWitness:               false,
				Quiesce:                 false,
			},
			InitMember: map[uint64]string{
				1: "0.0.0.0:9001",
				2: "0.0.0.0:9002",
				3: "0.0.0.0:9003",
			},
		},

		Directory: libndb.NutsDBFolder{
			Base:              "",
			Data:              "data",
			Backup:            "backup",
			Temp:              "temp",
			LimitNumberBackup: 5,
			Permission:        0770,
		},
	}

	return cfg
}
