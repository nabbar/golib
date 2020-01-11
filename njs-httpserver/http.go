/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package njs_httpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	njs_certif "github.com/nabbar/golib/njs-certif"
	njs_logger "github.com/nabbar/golib/njs-logger"
)

type modelServer struct {
	ssl  *tls.Config
	srv  *http.Server
	hdl  http.Handler
	addr *url.URL
	host string
	port int
}

type HTTPServer interface {
	GetBindable() string
	GetExpose() string
	IsRunning() bool

	Listen()
	Restart()
	Shutdown()

	WaitNotify()
}

func NewServer(listen, expose string, handler http.Handler, tlsConfig *tls.Config) HTTPServer {
	srv := &modelServer{
		hdl: handler,
		ssl: tlsConfig,
	}

	if host, prt, err := net.SplitHostPort(listen); err != nil {
		srv.host = listen
		srv.port = 0
	} else if port, err := strconv.Atoi(prt); err != nil {
		srv.host = host
		srv.port = 0
	} else {
		srv.host = host
		srv.port = port
	}

	if expose == "" {
		expose = listen
	}

	if uri, err := url.Parse(expose); err == nil {
		srv.addr = uri
	} else if uri, err = url.Parse(listen); err == nil {
		srv.addr = uri
	} else {
		srv.addr = &url.URL{
			Host: expose,
		}
	}

	if srv.addr.Scheme == "" {
		if srv.ssl != nil {
			srv.addr.Scheme = "https"
		} else {
			srv.addr.Scheme = "http"
		}
	} else if srv.addr.Scheme == "http" {
		srv.ssl = nil
	}

	return srv
}

func ListenWaitNotify(allSrv ...HTTPServer) {
	var wg sync.WaitGroup
	wg.Add(len(allSrv))

	for _, s := range allSrv {
		go func(serv HTTPServer) {
			defer wg.Done()
			serv.Listen()
			serv.WaitNotify()
		}(s)
	}

	wg.Wait()
}

func Listen(allSrv ...HTTPServer) {
	for _, s := range allSrv {
		go func(serv HTTPServer) {
			serv.Listen()
		}(s)
	}
}

func Restart(allSrv ...HTTPServer) {
	for _, s := range allSrv {
		s.Restart()
	}
}

func Shutdown(allSrv ...HTTPServer) {
	for _, s := range allSrv {
		s.Shutdown()
	}
}

func IsRunning(allSrv ...HTTPServer) bool {
	for _, s := range allSrv {
		if s.IsRunning() {
			return true
		}
	}

	return false
}

func (srv modelServer) GetBindable() string {
	return fmt.Sprintf("%s:%d", srv.host, srv.port)
}

func (srv modelServer) GetExpose() string {
	return srv.addr.String()
}

func (srv *modelServer) Listen() {
	if srv.srv != nil {
		srv.Shutdown()
	}

	srv.srv = &http.Server{
		Addr:      srv.GetBindable(),
		ErrorLog:  njs_logger.GetLogger(njs_logger.ErrorLevel, log.LstdFlags | log.LstdFlags | log.Lmicroseconds, "server '%s'", srv.GetBindable()),
		Handler:   srv.hdl,
		TLSConfig: srv.ssl,
	}

	njs_logger.InfoLevel.Logf("Server starting with bindable: %s", srv.GetBindable())

	go func() {
		if srv.ssl == nil || !njs_certif.CheckCertificates() {
			if err := srv.srv.ListenAndServe(); err != nil {
				njs_logger.FatalLevel.Logf("Listen Error: %v", err)
				return
			}
		} else {
			if err := srv.srv.ListenAndServeTLS("", ""); err != nil {
				njs_logger.FatalLevel.Logf("Listen config Error: %v", err)
				return
			}
		}
	}()
}

func (srv *modelServer) WaitNotify() {
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	//signal.Notify(quit, syscall.SIGKILL)
	signal.Notify(quit, syscall.SIGQUIT)

	<-quit
	srv.Shutdown()
}

func (srv *modelServer) Restart() {
	if srv.srv != nil {
		srv.Shutdown()
	}

	srv.Listen()
}

func (srv *modelServer) Shutdown() {
	njs_logger.InfoLevel.Logf("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if srv.srv == nil {
		return
	}

	if err := srv.srv.Shutdown(ctx); err != nil {
		njs_logger.FatalLevel.Logf("Server Shutdown Error: %v", err)
	}

	srv.srv = nil
}

func (srv *modelServer) IsRunning() bool {
	return srv.srv != nil
}
