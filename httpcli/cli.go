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
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	liberr "github.com/nabbar/golib/errors"
	libptc "github.com/nabbar/golib/network/protocol"
	"golang.org/x/net/http2"
)

const (
	ClientTimeout30Sec = 30 * time.Second
	ClientTimeout10Sec = 10 * time.Second
	ClientTimeout5Sec  = 5 * time.Second
	ClientTimeout1Sec  = 1 * time.Second

	ClientNetworkTCP = "tcp"
	ClientNetworkUDP = "udp"
)

var trp = new(atomic.Value)

func init() {
	trp.Store(&http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		TLSHandshakeTimeout:   10 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   1,
		MaxConnsPerHost:       25,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		},
	})
}

type FctHttpClient func() *http.Client

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func ForceUpdateTransport(cfg *tls.Config, proxyUrl *url.URL) *http.Transport {
	t := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		TLSClientConfig:       cfg.Clone(),
		TLSHandshakeTimeout:   10 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   1,
		MaxConnsPerHost:       25,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	if proxyUrl != nil {
		t.Proxy = http.ProxyURL(proxyUrl)
	}

	trp.Store(t)

	return t
}

func GetTransport(tlsConfig *tls.Config, proxyURL *url.URL, DisableKeepAlive, DisableCompression, ForceHTTP2 bool) *http.Transport {
	var tr *http.Transport

	if i := trp.Load(); i != nil {
		if t, k := i.(*http.Transport); k {
			tr = t.Clone()
		}
	}

	if tr == nil {
		tr = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			TLSHandshakeTimeout:   10 * time.Second,
			DisableKeepAlives:     false,
			DisableCompression:    false,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   1,
			MaxConnsPerHost:       25,
			IdleConnTimeout:       90 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
		}
	}

	tr.DisableCompression = DisableCompression
	tr.DisableKeepAlives = DisableKeepAlive
	tr.TLSClientConfig = tlsConfig.Clone()

	if proxyURL != nil {
		tr.Proxy = http.ProxyURL(proxyURL)
	}

	return tr
}

func SetTransportDial(tr *http.Transport, forceIp bool, netw libptc.NetworkProtocol, ip, local string) {
	var (
		dial = &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		fctDial func(ctx context.Context, network, address string) (net.Conn, error)
	)

	if forceIp && len(local) > 0 {
		u := &url.URL{
			Host: local,
		}
		fctDial = func(ctx context.Context, network, address string) (net.Conn, error) {
			dial.LocalAddr = &net.TCPAddr{
				IP: net.ParseIP(u.Hostname()),
			}

			return dial.DialContext(ctx, netw.Code(), ip)
		}
	} else if forceIp {
		fctDial = func(ctx context.Context, network, address string) (net.Conn, error) {
			return dial.DialContext(ctx, netw.Code(), ip)
		}
	} else {
		fctDial = dial.DialContext
	}

	tr.DialContext = fctDial
	tr.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return tls.DialWithDialer(dial, network, addr, tr.TLSClientConfig)
	}
}

func GetClient(tr *http.Transport, http2Tr bool, GlobalTimeout time.Duration) (*http.Client, liberr.Error) {
	if http2Tr {
		if e := http2.ConfigureTransport(tr); e != nil {
			return nil, ErrorClientTransportHttp2.Error(e)
		}
	}

	c := &http.Client{
		Transport: tr,
	}

	if GlobalTimeout != 0 {
		c.Timeout = GlobalTimeout
	}

	return c, nil
}
