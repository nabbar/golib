/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package httpcli

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	libdur "github.com/nabbar/golib/duration"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
)

const (
	ClientTimeout5Sec = 5 * time.Second
)

var (
	dns = new(atomic.Value)
	ctx context.Context
	cnl context.CancelFunc
)

func initDNSMapper() htcdns.DNSMapper {
	if cnl != nil {
		cnl()
	}

	ctx, cnl = context.WithCancel(context.Background())

	return htcdns.New(ctx, &htcdns.Config{
		DNSMapper:  make(map[string]string),
		TimerClean: libdur.ParseDuration(3 * time.Minute),
		Transport: htcdns.TransportConfig{
			Proxy:                 nil,
			TLSConfig:             &libtls.Config{},
			DisableHTTP2:          false,
			DisableKeepAlive:      false,
			DisableCompression:    false,
			MaxIdleConns:          50,
			MaxIdleConnsPerHost:   5,
			MaxConnsPerHost:       25,
			TimeoutGlobal:         libdur.ParseDuration(30 * time.Second),
			TimeoutKeepAlive:      libdur.ParseDuration(15 * time.Second),
			TimeoutTLSHandshake:   libdur.ParseDuration(10 * time.Second),
			TimeoutExpectContinue: libdur.ParseDuration(3 * time.Second),
			TimeoutIdleConn:       libdur.ParseDuration(30 * time.Second),
			TimeoutResponseHeader: 0,
		},
	}, nil)
}

func DefaultDNSMapper() htcdns.DNSMapper {
	if i := dns.Load(); i == nil {
		d := initDNSMapper()
		dns.Store(d)
		return d
	} else if d, k := i.(htcdns.DNSMapper); !k {
		d = initDNSMapper()
		dns.Store(d)
		return d
	} else {
		return d
	}
}

func SetDefaultDNSMapper(d htcdns.DNSMapper) {
	if d == nil {
		return
	}

	if cnl != nil {
		cnl()
	}

	dns.Store(d)
}

type FctHttpClient func() *http.Client
type FctHttpClientSrv func(servername string) *http.Client

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func GetClient() *http.Client {
	return DefaultDNSMapper().DefaultClient()
}
