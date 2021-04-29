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
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	libnat "github.com/nabbar/golib/nats"
	libpwd "github.com/nabbar/golib/password"
	libsem "github.com/nabbar/golib/semaphore"
	natcli "github.com/nats-io/nats.go"
)

const (
	BasePortServer  = 9000
	BasePortCluster = 9100
	BasePortHttp    = 9200
	BasePortProf    = 9300

	BasePathFolder = "/nats"
	SubPidFile     = "nats-node-%d.pid"
	SubPortDir     = "node-%d"
	SubLogFile     = "nats-node-%d.log"

	NbNodeInstance = 3
	NbEntries      = 1000000

	usrProducer  = "produser"
	pwdProducer  = "test123prod"
	usrSubsriber = "subuser"
	pwdSubsriber = "test123sub"
	usrCluster   = "cluster"
	pwdCluster   = "cLu!123-test"

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
			panic(err)
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
		go cluster[i].Listen(ctx)
	}

	cluster[0].WaitReady(ctx, 200*time.Millisecond)

	sem := libsem.NewSemaphoreWithContext(ctx, 0)

	defer sem.DeferMain()

	sub, err := cluster[1].Client(ctx, 200*time.Millisecond, nil, configClient("subs-test", usrSubsriber, pwdSubsriber))

	if err != nil {
		panic(err)
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

	prd, err := cluster[2].Client(ctx, 200*time.Millisecond, nil, configClient("prod-test", usrProducer, pwdProducer))
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

	time.Sleep(30 * time.Second)

	if e := s.Unsubscribe(); e != nil {
		panic(e)
	}

	prd.Close()
	time.Sleep(200 * time.Millisecond)
	sub.Close()
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
	//pidFile := fmt.Sprintf(filepath.Join(BasePathFolder, SubPidFile), id)
	//portDir := fmt.Sprintf(filepath.Join(BasePathFolder, SubPortDir), id)
	logFile := fmt.Sprintf(filepath.Join(BasePathFolder, SubLogFile), id)
	/*
		if _, err := os.Stat(portDir); err != nil && errors.Is(err, os.ErrNotExist) {
			if err = os.MkdirAll(portDir, 0755); err != nil {
				panic(err)
			}
		}
	*/
	if _, err := os.Stat(logFile); err != nil && errors.Is(err, os.ErrNotExist) {
		if h, e := os.Create(logFile); e != nil {
			panic(e)
		} else {
			_ = h.Close()
		}
	}

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
			Host:            "127.0.0.1",
			Port:            BasePortServer + id,
			ClientAdvertise: "", //fmt.Sprintf("127.0.0.1:%d", BasePortServer+id),
			HTTPHost:        "127.0.0.1",
			HTTPPort:        BasePortHttp + id,
			HTTPSPort:       0,
			ProfPort:        BasePortProf + id,
			//			PidFile:         pidFile,
			//			PortsFileDir:    portDir,
			Routes:     rts,
			RoutesStr:  "",
			NoSig:      false,
			TLS:        false,
			TLSTimeout: 0,
			TLSConfig: libtls.Config{
				InheritDefault: true,
			},
		},
		Cluster: libnat.ConfigCluster{
			Host:           "127.0.0.1",
			Port:           BasePortCluster + id,
			ListenStr:      "",
			Advertise:      fmt.Sprintf("127.0.0.1:%d", BasePortCluster+id),
			NoAdvertise:    false,
			ConnectRetries: 5,
			Username:       usrCluster,
			Password:       pwdCluster,
			AuthTimeout:    0,
			Permissions: libnat.ConfigPermissionsRoute{
				Import: libnat.ConfigPermissionSubject{
					Allow: []string{
						">",
					},
					Deny: make([]string, 0),
				},
				Export: libnat.ConfigPermissionSubject{
					Allow: []string{
						">",
					},
					Deny: make([]string, 0),
				},
			},
			TLS:        false,
			TLSTimeout: 0,
			TLSConfig: libtls.Config{
				InheritDefault: true,
			},
		},
		Limits: libnat.ConfigLimits{
			MaxConn:          0,
			MaxSubs:          0,
			PingInterval:     0,
			MaxPingsOut:      0,
			MaxControlLine:   0,
			MaxPayload:       0,
			MaxPending:       0,
			WriteDeadline:    0,
			RQSubsSweep:      0,
			MaxClosedClients: 0,
			LameDuckDuration: 0,
		},
		Logs: libnat.ConfigLogger{
			NoLog:        false,
			LogFile:      logFile,
			Syslog:       false,
			RemoteSyslog: "",
		},
		Auth: libnat.ConfigAuth{
			Users: []libnat.ConfigUser{
				{
					Username: usrProducer,
					Password: pwdProducer,
					Permissions: libnat.ConfigPermissionsUser{
						Publish: libnat.ConfigPermissionSubject{
							Allow: []string{
								Subject,
								Subject + ".*",
								Subject + ".>",
							},
							Deny: make([]string, 0),
						},
						Subscribe: libnat.ConfigPermissionSubject{
							Allow: []string{
								Subject,
								Subject + ".*",
								Subject + ".>",
							},
							Deny: make([]string, 0),
						},
					},
				},
				{
					Username: usrSubsriber,
					Password: pwdSubsriber,
					Permissions: libnat.ConfigPermissionsUser{
						Publish: libnat.ConfigPermissionSubject{
							Allow: []string{
								Subject,
								Subject + ".*",
								Subject + ".>",
							},
							Deny: make([]string, 0),
						},
						Subscribe: libnat.ConfigPermissionSubject{
							Allow: []string{
								Subject,
								Subject + ".*",
								Subject + ".>",
							},
							Deny: make([]string, 0),
						},
					},
				},
				{
					Username: usrCluster,
					Password: pwdCluster,
					Permissions: libnat.ConfigPermissionsUser{
						Publish: libnat.ConfigPermissionSubject{
							Allow: []string{
								">",
							},
							Deny: make([]string, 0),
						},
						Subscribe: libnat.ConfigPermissionSubject{
							Allow: []string{
								">",
							},
							Deny: make([]string, 0),
						},
					},
				},
			},
			AuthTimeout: 0,
		},
	}
}

func configClient(name, user, pass string) libnat.Client {
	return libnat.Client{
		Url:                         "",
		Servers:                     nil,
		NoRandomize:                 false,
		NoEcho:                      false,
		Name:                        name,
		Verbose:                     false,
		Pedantic:                    true,
		AllowReconnect:              true,
		MaxReconnect:                0,
		ReconnectWait:               0,
		ReconnectJitter:             0,
		ReconnectJitterTLS:          0,
		Timeout:                     0,
		DrainTimeout:                0,
		FlusherTimeout:              0,
		PingInterval:                0,
		MaxPingsOut:                 0,
		ReconnectBufSize:            0,
		SubChanLen:                  0,
		User:                        user,
		Password:                    pass,
		Token:                       "",
		UseOldRequestStyle:          false,
		NoCallbacksAfterClientClose: false,
		Secure:                      false,
		TLSConfig: libtls.Config{
			InheritDefault: true,
		},
	}
}

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
