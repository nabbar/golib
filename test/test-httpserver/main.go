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
)

func init() {
	liblog.EnableColor()
	liblog.SetLevel(liblog.DebugLevel)
	liblog.AddGID(true)
	liblog.FileTrace(true)
	liblog.SetFormat(liblog.TextFormat)
	liblog.Timestamp(true)

	liberr.SetModeReturnError(liberr.ErrorReturnCodeErrorTraceFull)

	ctx, cnl = context.WithCancel(context.Background())

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
		for {
			time.Sleep(5 * time.Second)
			if ctx.Err() != nil {
				return
			}
			pool.MapRun(func(srv libsrv.Server) {
				n, v, _ := srv.StatusInfo()
				if e := srv.StatusHealth(); e != nil {
					fmt.Printf("%s - %s : %v\n", n, v, e)
				} else {
					fmt.Printf("%s - %s : OK\n", n, v)
				}

			})
		}
	}()

	var i = 0
	for {

		time.Sleep(5 * time.Second)

		if ctx.Err() != nil {
			return
		}

		fmt.Printf("Srv Name : %v\n", pool.List(libsrv.FieldBind, libsrv.FieldName, "", ".*"))
		i++

		if i%3 == 0 {
			fmt.Printf("Reloading Server : %v\n", pool.List(libsrv.FieldBind, libsrv.FieldName, "", ".*"))
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
