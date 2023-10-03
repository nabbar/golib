//+build examples
//go:build examples
// +build examples

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
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/c-bata/go-prompt/completer"

	libclu "github.com/nabbar/golib/cluster"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	libndb "github.com/nabbar/golib/nutsdb"
	libpwd "github.com/nabbar/golib/password"
	libpgb "github.com/nabbar/golib/progress"
	libsh "github.com/nabbar/golib/shell"
	libvrs "github.com/nabbar/golib/version"
	"github.com/nutsdb/nutsdb"
	"github.com/vbauerster/mpb/v8"
)

const (
	BaseDirPattern = "/nutsdb/node-%d"
	NbInstances    = 3
	NbEntries      = 1000
	LoggerFile     = "/nutsdb/nutsdb.log"
	AllowPut       = false
	AllowGet       = true
)

var (
	bg  = new(atomic.Value)
	bp  = new(atomic.Value)
	log = new(atomic.Value)
	ctx context.Context
	cnl context.CancelFunc
)

func init() {
	ctx, cnl = context.WithCancel(context.Background())
	liberr.SetModeReturnError(liberr.ErrorReturnCodeErrorTraceFull)
	log.Store(liblog.New(ctx))
	initLogger()
}

type EmptyStruct struct{}

func main() {
	defer func() {
		if cnl != nil {
			cnl()
		}
	}()

	println(fmt.Sprintf("Running test with %d threads...", runtime.GOMAXPROCS(0)))
	println(fmt.Sprintf("Init cluster..."))
	tStart := time.Now()
	cluster := Start(ctx)

	liblog.SetLevel(liblog.WarnLevel)
	defer func() {
		Stop(ctx, cluster)
	}()

	if AllowPut || AllowGet {
		println(strings.Join(Inject(ctx, tStart, cluster), "\n"))
		os.Exit(0)
	}

	vs := libvrs.NewVersion(libvrs.License_MIT, "Test Raft DB Nuts", "Tools to test a raft cluster of nutsDB", time.Now().Format(time.RFC3339), "0000000", "v0.0.0", "No one", "pfx", EmptyStruct{}, 0)
	vs.PrintLicense()
	vs.PrintInfo()

	_, _ = fmt.Fprintf(os.Stdout, "Please use `exit` or `Ctrl-D` to exit this program.\n")

	sh := libsh.New()
	sh.Add("", cluster[0].ShellCommand(func() context.Context {
		return ctx
	}, 200*time.Millisecond)...)
	sh.RunPrompt(os.Stdout, os.Stderr,
		prompt.OptionTitle(fmt.Sprintf("%s: %s", vs.GetPackage(), vs.GetDescription())),
		prompt.OptionPrefix(">>> "),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
	)
}

func Inject(ctx context.Context, tStart time.Time, cluster []libndb.NutsDB) []string {
	lvl := liblog.GetCurrentLevel()
	liblog.SetLevel(liblog.WarnLevel)

	tInit := time.Since(tStart)
	mInit := fmt.Sprintf("Memory used after Init: \n%s", strings.Join(GetMemUsage(), "\n"))
	runtime.GC()
	println(fmt.Sprintf("Init done. \n"))

	pgb := libpgb.NewProgressBarWithContext(ctx, mpb.WithWidth(64), mpb.WithRefreshRate(200*time.Millisecond))
	barPut := pgb.NewBarSimpleCounter("PutEntry", int64(NbEntries))
	defer barPut.DeferMain(true)

	tStart = time.Now()
	for i := 0; i < NbEntries; i++ {
		if e := barPut.NewWorker(); e != nil {
			continue
		}

		go func(ctx context.Context, bar libpgb.Bar, clu libndb.NutsDB, num int) {
			defer bar.DeferWorker()
			if AllowPut {
				Put(ctx, clu, fmt.Sprintf("key-%03d", num), fmt.Sprintf("val-%03d|%s|%s|%s", num, libpwd.Generate(50), libpwd.Generate(50), libpwd.Generate(50)))
			}
		}(ctx, barPut, cluster[i%NbInstances], i+1)
	}

	if e := barPut.WaitAll(); e != nil {
		panic(e)
	}
	tPut := time.Since(tStart)
	mPut := fmt.Sprintf("Memory used after Put entries: \n%s", strings.Join(GetMemUsage(), "\n"))
	runtime.GC()

	barGet := pgb.NewBarSimpleCounter("GetEntry", int64(NbEntries))
	defer barGet.DeferMain(true)

	tStart = time.Now()
	for i := 0; i < NbEntries; i++ {
		if e := barGet.NewWorker(); e != nil {
			continue
		}

		c := i%NbInstances + 1
		if c == NbInstances {
			c = 0
		}

		go func(ctx context.Context, bar libpgb.Bar, clu libndb.NutsDB, num int) {
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
	liblog.SetLevel(liblog.InfoLevel)
	liblog.InfoLevel.Logf("Results testing: \n%s", strings.Join(res, "\n"))
	liblog.SetLevel(lvl)
	return res
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
		liblog.ErrorLevel.Logf("Cmd Get for key '%s', error : %v", key, e)
		fmt.Printf("Cmd Get for key '%s', error : %v", key, e)
	} else if !bytes.HasPrefix(v.Value, []byte(val)) {
		liblog.ErrorLevel.Logf("Cmd Get for key '%s', awaiting value start with '%s', but find : %s", key, val, string(v.Value))
		fmt.Printf("Cmd Get for key '%s', awaiting value start with '%s', but find : %s", key, val, string(v.Value))
	}
}

func Start(ctx context.Context) []libndb.NutsDB {
	var clusters = make([]libndb.NutsDB, NbInstances)

	for i := 0; i < NbInstances; i++ {
		clusters[i] = initNutDB(i + 1)
		clusters[i].SetLogger(func() liblog.Logger {
			l := getLogger()
			l.SetFields(l.GetFields().Add("lib", libndb.LogLib).Add("instance", i))
			return l
		})

		liblog.InfoLevel.Logf("Starting node ID #%d...", i+1)
		if err := clusters[i].Listen(); err != nil {
			panic(err)
		}

		time.Sleep(5 * time.Second)
	}

	return clusters
}

func Stop(ctx context.Context, clusters []libndb.NutsDB) {
	for i := 0; i < NbInstances; i++ {
		liblog.InfoLevel.Logf("Stopping node ID #%d...", i+1)
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
				4: "localhost:9004",
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

func getLogger() liblog.Logger {
	if log == nil {
		return liblog.New(context.Background())
	} else if i := log.Load(); i == nil {
		return liblog.New(context.Background())
	} else if l, ok := i.(liblog.Logger); !ok {
		return liblog.New(context.Background())
	} else {
		return l
	}
}

func getLoggerDgb() liblog.Logger {
	l := getLogger()
	l.SetFields(l.GetFields().Add("lib", libclu.LogLib))
	return l
}

func setLogger(l liblog.Logger) {
	if log == nil {
		log = new(atomic.Value)
	}

	log.Store(l)
}

func initLogger() {
	l := getLogger()
	l.SetLevel(liblog.InfoLevel)
	if err := l.SetOptions(&liblog.Options{
		DisableStandard:  true,
		DisableStack:     false,
		DisableTimestamp: false,
		EnableTrace:      false,
		TraceFilter:      "",
		DisableColor:     false,
		LogFile: []liblog.OptionsFile{
			{
				LogLevel: []string{
					"panic",
					"fatal",
					"error",
					"warning",
					"info",
					"debug",
				},
				Filepath:         LoggerFile,
				Create:           true,
				CreatePath:       true,
				FileMode:         0644,
				PathMode:         0755,
				DisableStack:     false,
				DisableTimestamp: false,
				EnableTrace:      true,
			},
		},
	}); err != nil {
		panic(err)
	}

	setLogger(l)
	libclu.SetLoggerFactory(getLoggerDgb)
}
