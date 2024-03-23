/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	iotfds "github.com/nabbar/golib/ioutils/fileDescriptor"
	libptc "github.com/nabbar/golib/network/protocol"
	libpgo "github.com/nabbar/golib/pprof"
	libsem "github.com/nabbar/golib/semaphore"
	libsiz "github.com/nabbar/golib/size"
	"golang.org/x/net/http2"
	"golang.org/x/sys/cpu"
)

func GetFreePort() int {
	var (
		addr *net.TCPAddr
		lstn *net.TCPListener
		err  error
	)

	if addr, err = net.ResolveTCPAddr("tcp", "localhost:0"); err != nil {
		panic(err)
	}

	if lstn, err = net.ListenTCP("tcp", addr); err != nil {
		panic(err)
	}

	defer func() {
		_ = lstn.Close()
	}()

	return lstn.Addr().(*net.TCPAddr).Port
}

func benchCurl(show bool, cli *http.Client, addr string) {
	var (
		e error
		r *http.Response
		t = time.Now()
		b []byte
	)

	if r, e = cli.Get("https://" + addr); e != nil {
		panic(e)
	}

	defer func() {
		_ = r.Body.Close()
	}()

	if b, e = io.ReadAll(r.Body); e != nil {
		panic(e)
	} else if show {
		_, _ = fmt.Fprintf(os.Stdout, "GET https://%s - %s - %dµs\n", addr, libsiz.Size(int64(len(b))).String(), time.Since(t).Microseconds())
	}
}

func serverListening(addr string) {
	var (
		err error
		srv *http.Server
		sv2 = &http2.Server{}

		cert tls.Certificate
	)

	if cert, err = genTLSCertificate(true); err != nil {
		panic(err)
	} else {
		srv = &http.Server{
			Addr:    addr,
			Handler: http.DefaultServeMux,
			TLSConfig: &tls.Config{
				Certificates:     []tls.Certificate{cert},
				MinVersion:       tls.VersionTLS13,
				MaxVersion:       tls.VersionTLS13,
				CurvePreferences: append(make([]tls.CurveID, 0), tls.X25519),
				CipherSuites:     append(make([]uint16, 0), tls.TLS_CHACHA20_POLY1305_SHA256),
			},
			ReadHeaderTimeout: 30 * time.Second,
			IdleTimeout:       time.Minute,
		}

		srv.SetKeepAlivesEnabled(true)
		if err = http2.ConfigureServer(srv, sv2); err != nil {
			panic(err)
		}
	}

	if e := srv.ListenAndServeTLS("", ""); e != nil {
		panic(e)
	}
}

func waitRunning(ctx context.Context, addr string) {
	for {
		d := &net.Dialer{}
		if co, ce := d.DialContext(ctx, libptc.NetworkTCP.Code(), addr); ce != nil {
			continue
		} else {
			_ = co.Close()
			return
		}
	}
}

func main() {
	if os.Getenv("PPROF") != "" {
		libpgo.ProfilingCPUStart()
		defer libpgo.ProfilingCPUDefer()
	}

	testing.Init()
	flag.Parse()

	var (
		ctx, cnl = context.WithCancel(context.Background())
		port     = GetFreePort()
		addr     = fmt.Sprintf(":%d", port)
	)

	defer cnl()

	_, _ = fmt.Fprintf(os.Stdout, "AVX: \t\t%v\n", cpu.X86.HasAVX)
	_, _ = fmt.Fprintf(os.Stdout, "AVX2: \t\t%v\n", cpu.X86.HasAVX2)
	_, _ = fmt.Fprintf(os.Stdout, "AVX512: \t%v\n", cpu.X86.HasAVX512)

	RunInit(ctx, addr)

	res := testing.Benchmark(func(b *testing.B) {
		RunQuery(ctx, addr, b)
	})

	_, _ = fmt.Fprintf(os.Stdout, "Benchmark: \n")
	_, _ = fmt.Fprintf(os.Stdout, "\tNumber: \t%d\n", res.N)
	_, _ = fmt.Fprintf(os.Stdout, "\tPerf: \t\t%0.3f µs/curl\n", float64(res.NsPerOp())/1000)
	_, _ = fmt.Fprintf(os.Stdout, "\tDuration: \t%s\n", res.T.String())
	_, _ = fmt.Fprintf(os.Stdout, "Memory: \n")
	_, _ = fmt.Fprintf(os.Stdout, "\tAllocs: \t%s\n", libsiz.Size(res.MemAllocs).String())
	_, _ = fmt.Fprintf(os.Stdout, "\tBytes: \t\t%s\n", libsiz.Size(res.MemBytes).String())
}

func RunInit(ctx context.Context, addr string) {
	if _, _, err := iotfds.SystemFileDescriptor(1024 * 1024); err != nil {
		panic(err)
	}

	go serverListening(addr)
	_, _ = fmt.Fprintf(os.Stdout, "Server Listenning '%s'...\n", addr)

	waitRunning(ctx, addr) // wait for server to start
}

func RunQuery(ctx context.Context, addr string, b *testing.B) {
	var sem = libsem.New(ctx, 0, false)
	defer sem.DeferMain()

	var cli = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	for i := 0; i < b.N; i++ {
		if err := sem.NewWorker(); err != nil {
			panic(err)
		}
		go func() {
			defer func() {
				if e := recover(); e != nil {
					_, _ = fmt.Fprintf(os.Stdout, "panic: %v\n", e)
				}
				sem.DeferWorker()
			}()
			benchCurl(false, cli, addr)
		}()
	}

	if err := sem.WaitAll(); err != nil {
		panic(err)
	}
}
