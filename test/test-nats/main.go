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
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	libnat "github.com/nabbar/golib/nats"
	libpwd "github.com/nabbar/golib/password"
	libsem "github.com/nabbar/golib/semaphore"
	"github.com/nats-io/jwt/v2"
	natcli "github.com/nats-io/nats.go"
)

const (
	BasePortServer  = 9000
	BasePortCluster = 9100
	BasePortHttp    = 9200
	BasePortProf    = 9300

	BasePathFolder = "/nats"
	SubLogFile     = "nats-node-%d.log"
	SubNodeDir     = "node-%d"

	NbNodeInstance = 3
	NbEntries      = 1000000

	nameProducer  = "produser"
	nameSubsriber = "subuser"
	cptCluster    = "cluster"
	usrCluster    = "cluster"
	pwdCluster    = "cLu!123-test"

	Subject = "_INBOX"
)

var (
	cluster []libnat.Server
	ctx     context.Context
	cnl     context.CancelFunc
	rng     *rand.Rand
)

type Message struct {
	Id     int
	Random []string
}

func init() {
	liblog.SetLevel(liblog.InfoLevel)
	liblog.EnableColor()
	liblog.AddGID(true)
	liblog.FileTrace(false)
	liberr.SetModeReturnError(liberr.ErrorReturnCodeErrorTrace)

	cluster = make([]libnat.Server, NbNodeInstance)
	rng = rand.New(&cryptoSource{})
}

func main() {
	ctx, cnl = context.WithCancel(context.Background())
	defer cnl()

	for i := 0; i < NbNodeInstance; i++ {
		cfg := configServer(i)
		if opt, err := cfg.NatsOption(nil); err != nil {
			panic(err.GetErrorFull(""))
		} else if err = cfg.LogConfigJson(); err != nil {
			panic(err.CodeErrorTraceFull("", ""))
		} else {
			cluster[i] = libnat.NewServer(opt)
		}
	}

	defer func() {
		for i := 0; i < NbNodeInstance; i++ {
			if cluster[i] != nil {
				cluster[i].Shutdown()
			}
		}
	}()

	for i := 0; i < NbNodeInstance; i++ {
		go func(ctx context.Context, clu libnat.Server) {
			if e := clu.Listen(ctx); e != nil {
				panic(e.CodeErrorTraceFull("", ""))
			}
		}(ctx, cluster[i])
	}

	cluster[0].WaitReady(ctx, 200*time.Millisecond)

	sem := libsem.NewSemaphoreWithContext(ctx, 0)

	defer sem.DeferMain()

	optSub := configClient(nameSubsriber, usrCluster, pwdCluster)
	sub, err := optSub.NewClient(nil)

	if err != nil {
		panic(err.CodeErrorTraceFull("", ""))
	}

	defer func() {
		if sub != nil && !sub.IsClosed() {
			sub.Close()
		}
	}()

	s, e := sub.Subscribe(Subject, func(msg *natcli.Msg) {
		if !msg.Sub.IsValid() {
			return
		}

		go func(m *natcli.Msg) {
			id, random := Unserialize(m.Data)
			r := time.Duration(rng.Intn(1000)) * time.Millisecond
			time.Sleep(r)
			_, _ = fmt.Fprintf(os.Stdout, "Subscriber read id '%06d' (after %s wait) : %s\n", id, r.String(), strings.Join(random, "|"))
		}(msg)
	})

	if e != nil {
		panic(e)
	}

	defer func() {
		if s == nil {
			return
		}
		if !s.IsValid() {
			return
		}
		if e = s.Unsubscribe(); e != nil {
			panic(e)
		}
	}()

	if !s.IsValid() {
		panic(s)
	}

	time.Sleep(3 * time.Second)
	optPrd := configClient(nameProducer, usrCluster, pwdCluster)
	prd, err := optPrd.NewClient(nil)
	if err != nil {
		panic(err)
	}

	defer func() {
		if prd != nil && !prd.IsClosed() {
			prd.Close()
		}
	}()

	for i := 0; i < NbEntries; i++ {
		msg := []string{
			libpwd.Generate(64),
			libpwd.Generate(64),
			libpwd.Generate(64),
			libpwd.Generate(64),
		}

		if err = sem.NewWorker(); err != nil {
			panic(err)
		}

		go func(idx int, msg []string, sem libsem.Sem) {
			defer sem.DeferWorker()

			time.Sleep(500 * time.Microsecond)

			if e := prd.Publish(Subject, Serialize(idx, msg)); e != nil {
				panic(e)
			}

			_, _ = fmt.Fprintf(os.Stdout, "Publisher write id '%06d' : %s\n", idx, strings.Join(msg, "|"))
		}(i, msg, sem)
	}

	if e := sem.WaitAll(); e != nil {
		panic(e)
	}

	time.Sleep(10 * time.Second)

	if e := s.Unsubscribe(); e != nil {
		panic(e)
	}

	prd.Close()
	time.Sleep(200 * time.Millisecond)
	sub.Close()

	println(strings.Join(GetMemUsage(), "\n"))
}

func Serialize(idx int, random []string) []byte {
	msg := Message{
		Id:     idx,
		Random: random,
	}

	if r, e := cbor.Marshal(msg); e != nil {
		panic(e)
	} else {
		return r
	}
}

func Unserialize(p []byte) (idx int, random []string) {
	msg := Message{}

	if e := cbor.Unmarshal(p, &msg); e != nil {
		panic(e)
	} else {
		return msg.Id, msg.Random
	}
}

func configServer(id int) libnat.Config {
	rts := make([]*url.URL, 0)

	for j := 0; j < NbNodeInstance; j++ {
		if j == id {
			continue
		}

		rts = append(rts, &url.URL{
			Scheme:      "nats",
			Opaque:      "",
			User:        url.UserPassword(usrCluster, pwdCluster),
			Host:        fmt.Sprintf("127.0.0.1:%d", BasePortCluster+j),
			Path:        "",
			RawPath:     "",
			ForceQuery:  false,
			RawQuery:    "",
			Fragment:    "",
			RawFragment: "",
		})
	}

	return libnat.Config{
		Server: libnat.ConfigSrv{
			Name:               fmt.Sprintf("node-%d", id),
			Host:               "127.0.0.1",
			Port:               BasePortServer + id,
			HTTPHost:           "127.0.0.1",
			HTTPPort:           BasePortHttp + id,
			ProfPort:           BasePortProf + id,
			Routes:             rts,
			NoSig:              true,
			JetStream:          true,
			StoreDir:           fmt.Sprintf(filepath.Join(BasePathFolder, SubNodeDir), id),
			PermissionStoreDir: 0755,
			TLS:                false,
			AllowNoTLS:         true,
			TLSConfig: libtls.Config{
				InheritDefault: true,
			},
		},
		Cluster: libnat.ConfigCluster{
			Name:           "Test-cluster",
			Host:           "127.0.0.1",
			Port:           BasePortCluster + id,
			ConnectRetries: 5,
			TLS:            false,
			TLSConfig: libtls.Config{
				InheritDefault: true,
			},
		},
		Gateways:   libnat.ConfigGateway{},
		Leaf:       libnat.ConfigLeaf{},
		Websockets: libnat.ConfigWebsocket{},
		MQTT:       libnat.ConfigMQTT{},
		Limits:     libnat.ConfigLimits{},
		Logs: libnat.ConfigLogger{
			LogFile:                 fmt.Sprintf(filepath.Join(BasePathFolder, SubLogFile), id),
			PermissionFolderLogFile: 0755,
			PermissionFileLogFile:   0644,
			Syslog:                  false,
		},
		Auth: libnat.ConfigAuth{
			NKeys: nil,
			Users: []libnat.ConfigUser{
				{
					Username: usrCluster,
					Password: pwdCluster,
					Account:  cptCluster,
					AllowedConnectionTypes: []string{
						jwt.ConnectionTypeStandard,
						jwt.ConnectionTypeLeafnode,
						jwt.ConnectionTypeWebsocket,
						jwt.ConnectionTypeMqtt,
					},
				},
			},
			Accounts: []libnat.ConfigAccount{
				{
					Name: cptCluster,
					Permission: libnat.ConfigPermissionsUser{
						Publish: libnat.ConfigPermissionSubject{
							Allow: []string{
								">",
								"*",
							},
							Deny: make([]string, 0),
						},
						Subscribe: libnat.ConfigPermissionSubject{
							Allow: []string{
								">",
								"*",
							},
							Deny: make([]string, 0),
						},
						Response: libnat.ConfigPermissionResponse{
							MaxMsgs: NbEntries,
							Expires: time.Second,
						},
					},
				},
			},
			SystemAccount:    cptCluster,
			NoSystemAccount:  false,
			AllowNewAccounts: true,
			TrustedKeys:      make([]string, 0),
			TrustedOperators: make([]string, 0),
		},
		Customs: nil,
	}
}

func configClient(name, user, pass string) libnat.Client {
	var srv = make([]string, 0)

	for i := 0; i < NbNodeInstance; i++ {
		srv = append(srv, fmt.Sprintf("nats://127.0.0.1:%d", BasePortServer+i))
	}

	return libnat.Client{
		Name:           name,
		Servers:        srv,
		Pedantic:       false,
		AllowReconnect: true,
		User:           user,
		Password:       pass,
		TLSConfig: libtls.Config{
			InheritDefault: true,
		},
	}
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

// random functions

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoSource) Uint64() (v uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}
