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
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/nabbar/golib/logger"

	libclu "github.com/nabbar/golib/cluster"
	liberr "github.com/nabbar/golib/errors"
	libndb "github.com/nabbar/golib/nutsdb"
	"github.com/nabbar/golib/progress"
	"github.com/vbauerster/mpb/v5"
	"github.com/xujiajun/nutsdb"
)

const (
	BaseDirPattern = "/nutsdb/node-%d"
	NbInstances    = 3
	NbEntries      = 100000
	LoggerFile     = "/nutsdb/nutsdb.log"
)

var (
	bg = new(atomic.Value)
	bp = new(atomic.Value)
)

func init() {
	liberr.SetModeReturnError(liberr.ErrorReturnCodeErrorTraceFull)
	logger.SetLevel(logger.WarnLevel)
	logger.AddGID(true)
	logger.EnableColor()
	logger.FileTrace(true)
	logger.Timestamp(true)
}

func main() {
	if _, err := os.Stat(LoggerFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	} else if err == nil {
		if err = os.Remove(LoggerFile); err != nil {
			panic(err)
		}
	}

	//nolint #gosec
	/* #nosec */
	if file, err := os.OpenFile(LoggerFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		panic(err)
	} else {
		logger.SetOutput(file)
		defer func() {
			if file != nil {
				logger.SetOutput(os.Stdout)
				_ = file.Close()
			}
		}()
	}

	ctx, cnl := context.WithCancel(context.Background())
	defer func() {
		if cnl != nil {
			cnl()
		}
	}()

	println(fmt.Sprintf("Init cluster..."))
	tStart := time.Now()
	cluster := Start(ctx)
	tInit := time.Since(tStart)
	println(fmt.Sprintf("Init done. \n"))

	pgb := progress.NewProgressBarWithContext(ctx, mpb.WithWidth(64), mpb.WithRefreshRate(200*time.Millisecond))
	barPut := pgb.NewBarSimpleCounter("PutEntry", int64(NbEntries))
	defer barPut.DeferMain(false)

	tStart = time.Now()
	for i := 0; i < NbEntries; i++ {
		if e := barPut.NewWorker(); e != nil {
			continue
		}

		go func(ctx context.Context, bar progress.Bar, clu libndb.NutsDB, num int) {
			defer bar.DeferWorker()
			Put(ctx, clu, fmt.Sprintf("key-%3d", num), fmt.Sprintf("val-%03d", num))
		}(ctx, barPut, cluster[i%3], i)
	}

	if e := barPut.WaitAll(); e != nil {
		panic(e)
	}
	tPut := time.Since(tStart)

	barGet := pgb.NewBarSimpleCounter("GetEntry", int64(NbEntries))
	defer barGet.DeferMain(false)

	tStart = time.Now()
	for i := 0; i < NbEntries; i++ {
		if e := barGet.NewWorker(); e != nil {
			continue
		}

		c := i%3 + 1
		if c == 3 {
			c = 0
		}

		go func(ctx context.Context, bar progress.Bar, clu libndb.NutsDB, num int) {
			defer bar.DeferWorker()
			Get(ctx, clu, fmt.Sprintf("key-%3d", num))
		}(ctx, barGet, cluster[c], i)
	}

	if e := barGet.WaitAll(); e != nil {
		panic(e)
	}
	tGet := time.Since(tStart)

	time.Sleep(10 * time.Second)

	println(fmt.Sprintf("Time for init cluster: %s", tInit.String()))
	println(fmt.Sprintf("Time for %d Put in DB: %s", NbEntries, tPut.String()))
	println(fmt.Sprintf("Time for %d Get in DB: %s", NbEntries, tGet.String()))
}

func Put(ctx context.Context, c libndb.NutsDB, key, val string) {
	_ = c.Client(ctx, 100*time.Microsecond).Put("myBucket", []byte(key), []byte(val), 0)
	//res := c.Client(ctx, 100*time.Microsecond).Put("myBucket", []byte(key), []byte(val), 0)
	//fmt.Printf("Cmd Put(%s, %s) : %v\n", key, val, res)
}

func Get(ctx context.Context, c libndb.NutsDB, key string) {
	_, _ = c.Client(ctx, 100*time.Microsecond).Get("myBucket", []byte(key))
	//v, e := c.Client(ctx, 100*time.Microsecond).Get("myBucket", []byte(key))
	//fmt.Printf("Cmd Get(%s) : %v --- err : %v\n", key, string(v.Value), e)
}

func Start(ctx context.Context) []libndb.NutsDB {
	var clusters = make([]libndb.NutsDB, NbInstances)

	for i := 0; i < NbInstances; i++ {
		clusters[i] = initNutDB(i + 1)

		if err := clusters[i].Listen(); err != nil {
			panic(err)
		}

		time.Sleep(5 * time.Second)
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
				RTTMillisecond:                200,
				RaftAddress:                   "",
				AddressByNodeHostID:           false,
				ListenAddress:                 "",
				MutualTLS:                     false,
				CAFile:                        "",
				CertFile:                      "",
				KeyFile:                       "",
				EnableMetrics:                 true,
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
				ElectionRTT:             15,
				HeartbeatRTT:            1,
				SnapshotEntries:         10,
				CompactionOverhead:      0,
				OrderedConfigChange:     false,
				MaxInMemLogSize:         0,
				SnapshotCompressionType: 0,
				EntryCompressionType:    0,
				DisableAutoCompactions:  true,
				IsObserver:              false,
				IsWitness:               false,
				Quiesce:                 false,
			},
			InitMember: map[uint64]string{
				1: "localhost:9001",
				2: "localhost:9002",
				3: "localhost:9003",
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
