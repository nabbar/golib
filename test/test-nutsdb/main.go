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
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nabbar/golib/password"

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
	NbEntries      = 10000
	LoggerFile     = "/nutsdb/nutsdb.log"
	AllowPut       = true
	AllowGet       = true
)

var (
	bg = new(atomic.Value)
	bp = new(atomic.Value)
)

func init() {
	liberr.SetModeReturnError(liberr.ErrorReturnCodeErrorTraceFull)
	logger.SetLevel(logger.InfoLevel)
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

	println(fmt.Sprintf("Running test with %d threads...", runtime.GOMAXPROCS(0)))
	println(fmt.Sprintf("Init cluster..."))
	tStart := time.Now()
	cluster := Start(ctx)

	logger.SetLevel(logger.WarnLevel)
	defer func() {
		Stop(ctx, cluster)
	}()

	tInit := time.Since(tStart)
	mInit := fmt.Sprintf("Memory used after Init: \n%s", strings.Join(GetMemUsage(), "\n"))
	runtime.GC()
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
			if AllowPut {
				Put(ctx, clu, fmt.Sprintf("key-%03d", num), fmt.Sprintf("val-%03d|%s|%s|%s", num, password.Generate(50), password.Generate(50), password.Generate(50)))
			}
		}(ctx, barPut, cluster[i%3], i+1)
	}

	if e := barPut.WaitAll(); e != nil {
		panic(e)
	}
	tPut := time.Since(tStart)
	mPut := fmt.Sprintf("Memory used after Put entries: \n%s", strings.Join(GetMemUsage(), "\n"))
	runtime.GC()

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
			if AllowGet {
				Get(ctx, clu, fmt.Sprintf("key-%03d", num), fmt.Sprintf("val-%03d", num))
			}
		}(ctx, barGet, cluster[c], i+1)
	}

	if e := barGet.WaitAll(); e != nil {
		panic(e)
	}
	tGet := time.Since(tStart)
	mGet := fmt.Sprintf("Memory used after Get entries: \n%s", strings.Join(GetMemUsage(), "\n"))
	runtime.GC()

	barPut.DeferMain(false)
	barPut = nil

	barGet.DeferMain(false)
	barGet = nil

	pgb = nil
	res := []string{
		fmt.Sprintf("Time for init cluster: %s", tInit.String()),
		fmt.Sprintf("Time for %d Put in DB: %s ( %s by entry )", NbEntries, tPut.String(), (tPut / NbEntries).String()),
		fmt.Sprintf("Time for %d Get in DB: %s ( %s by entry )", NbEntries, tGet.String(), (tGet / NbEntries).String()),
		mInit,
		mPut,
		mGet,
	}
	runtime.GC()
	logger.SetLevel(logger.InfoLevel)
	time.Sleep(5 * time.Second)

	println(strings.Join(res, "\n"))
	logger.InfoLevel.Logf("Results testing: \n%s", strings.Join(res, "\n"))
	time.Sleep(5 * time.Second)
}

func GetMemUsage() []string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return []string{
		fmt.Sprintf("\t - Alloc      = %v MiB", m.Alloc/1024/1024),
		fmt.Sprintf("\t - TotalAlloc = %v MiB", m.TotalAlloc/1024/1024),
		fmt.Sprintf("\t - Sys        = %v MiB", m.Sys/1024/1024),
		fmt.Sprintf("\t - NumGC      = %v\n", m.NumGC),
	}
}

func Put(ctx context.Context, c libndb.NutsDB, key, val string) {
	_ = c.Client(ctx, 100*time.Microsecond).Put("myBucket", []byte(key), []byte(val), 0)
	//res := c.Client(ctx, 100*time.Microsecond).Put("myBucket", []byte(key), []byte(val), 0)
	//fmt.Printf("Cmd Put(%s, %s) : %v\n", key, val, res)
}

func Get(ctx context.Context, c libndb.NutsDB, key, val string) {
	//_, _ = c.Client(ctx, 100*time.Microsecond).Get("myBucket", []byte(key))
	v, e := c.Client(ctx, 100*time.Microsecond).Get("myBucket", []byte(key))
	if e != nil {
		logger.ErrorLevel.Logf("Cmd Get for key '%s', error : %v", key, e)
		fmt.Printf("Cmd Get for key '%s', error : %v", key, e)
	} else if !bytes.HasPrefix(v.Value, []byte(val)) {
		logger.ErrorLevel.Logf("Cmd Get for key '%s', awaiting value start with '%s', but find : %s", key, val, string(v.Value))
		fmt.Printf("Cmd Get for key '%s', awaiting value start with '%s', but find : %s", key, val, string(v.Value))
	}
}

func Start(ctx context.Context) []libndb.NutsDB {
	var clusters = make([]libndb.NutsDB, NbInstances)

	for i := 0; i < NbInstances; i++ {
		clusters[i] = initNutDB(i + 1)

		logger.InfoLevel.Logf("Starting node ID #%d...", i+1)
		if err := clusters[i].Listen(); err != nil {
			panic(err)
		}

		time.Sleep(5 * time.Second)
	}

	return clusters
}

func Stop(ctx context.Context, clusters []libndb.NutsDB) {
	for i := 0; i < NbInstances; i++ {
		logger.InfoLevel.Logf("Stopping node ID #%d...", i+1)
		if err := clusters[i].Shutdown(); err != nil {
			panic(err)
		}

		time.Sleep(5 * time.Second)
	}
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
			EntryIdxMode:         nutsdb.HintKeyAndRAMIdxMode,
			RWMode:               nutsdb.FileIO,
			SegmentSize:          64 * 1024,
			SyncEnable:           true,
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
