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
	"net/http"
	"time"

	liblog "github.com/nabbar/golib/logger"

	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	libsrv "github.com/nabbar/golib/httpserver"
)

var tlsConfigSrv = libtls.Config{
	InheritDefault: true,
	VersionMin:     "1.2",
}

var cfgSrv01 = libsrv.ServerConfig{
	Name:         "test-01",
	Listen:       "0.0.0.0:61001",
	Expose:       "0.0.0.0:61000",
	TLSMandatory: false,
	TLS:          tlsConfigSrv,
}

var cfgSrv02 = libsrv.ServerConfig{
	Name:         "test-02",
	Listen:       "0.0.0.0:61002",
	Expose:       "0.0.0.0:61000",
	TLSMandatory: false,
	TLS:          tlsConfigSrv,
}

var cfgSrv03 = libsrv.ServerConfig{
	Name:         "test-03",
	Listen:       "0.0.0.0:61003",
	Expose:       "0.0.0.0:61000",
	TLSMandatory: true,
	TLS:          tlsConfigSrv,
}

var (
	cfgPool libsrv.PoolServerConfig
	ctx     context.Context
	cnl     context.CancelFunc
	log     = liblog.New()
)

func init() {
	liberr.SetModeReturnError(liberr.ErrorReturnCodeErrorTraceFull)

	ctx, cnl = context.WithCancel(context.Background())

	log.SetLevel(liblog.DebugLevel)
	if err := log.SetOptions(ctx, &liblog.Options{
		DisableStandard:  false,
		DisableStack:     false,
		DisableTimestamp: false,
		EnableTrace:      true,
		TraceFilter:      "",
		DisableColor:     false,
		LogSyslog: []liblog.OptionsSyslog{
			{
				LogLevel:         nil,
				Network:          "",
				Host:             "",
				Severity:         "",
				Facility:         "USER",
				Tag:              "test-http-server",
				DisableStack:     false,
				DisableTimestamp: false,
				EnableTrace:      true,
			},
		},
	}); err != nil {
		panic(err)
	}

	cfgPool = libsrv.PoolServerConfig{cfgSrv01, cfgSrv02, cfgSrv03}
	cfgPool.MapUpdate(func(cfg libsrv.ServerConfig) libsrv.ServerConfig {
		cfg.SetParentContext(func() context.Context {
			return ctx
		})
		return cfg
	})

}

func main() {
	var (
		pool libsrv.PoolServer
		lerr liberr.Error
	)

	if pool, lerr = cfgPool.PoolServer(); lerr != nil {
		panic(lerr)
	} else {
		pool.SetLogger(getLogger)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)
	mux.HandleFunc("/headers", headers)

	if lerr = pool.Listen(mux); lerr != nil {
		panic(lerr)
	}

	defer func() {
		pool.Shutdown()
		cnl()
	}()

	go pool.WaitNotify(ctx, cnl)

	go func() {
		l := getLogger()
		for {
			time.Sleep(5 * time.Second)

			if ctx.Err() != nil {
				return
			}

			pool.MapRun(func(srv libsrv.Server) {
				n, v, _ := srv.StatusInfo()
				e := l.Entry(liblog.ErrorLevel, "status message")
				e = e.FieldAdd("server_name", n).FieldAdd("server_release", v)
				e.ErrorAdd(true, srv.StatusHealth())
				e.Check(liblog.InfoLevel)
			})
		}
	}()

	var i = 0
	for {

		l := getLogger()
		time.Sleep(5 * time.Second)

		if ctx.Err() != nil {
			return
		}

		i++

		if i%3 == 0 {
			for s := range pool.List(libsrv.FieldBind, libsrv.FieldName, "", ".*") {
				l.Entry(liblog.InfoLevel, "Restarting server...").FieldAdd("server_name", s).Log()
			}
			pool.Restart()
			i = 0
		}
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "hello\n")
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			_, _ = fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func getLogger() liblog.Logger {
	l, _ := log.Clone(ctx)
	return l
}
